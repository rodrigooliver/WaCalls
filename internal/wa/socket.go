package wa

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"wacalls/internal/voip/core"
	"wacalls/internal/voip/signaling"

	"go.mau.fi/whatsmeow"
	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
)

type Socket struct {
	cli *whatsmeow.Client
}

func NewSocket(cli *whatsmeow.Client) *Socket { return &Socket{cli: cli} }

var _ core.VoipSocket = (*Socket)(nil)

func (s *Socket) di() *whatsmeow.DangerousInternalClient { return s.cli.DangerousInternals() }

func (s *Socket) OwnPN() types.JID { return s.di().GetOwnID() }

func (s *Socket) OwnLID() types.JID { return s.di().GetOwnLID() }

func (s *Socket) AccountDeviceIdentityNode() (waBinary.Node, bool) {
	if s.cli.Store == nil || s.cli.Store.Account == nil {
		return waBinary.Node{}, false
	}
	return s.di().MakeDeviceIdentityNode(), true
}

func (s *Socket) SendNode(ctx context.Context, node waBinary.Node) error {
	return s.di().SendNode(ctx, node)
}

func (s *Socket) Query(ctx context.Context, node waBinary.Node) (*waBinary.Node, error) {
	id, _ := node.Attrs["id"].(string)
	if id == "" {
		return nil, s.di().SendNode(ctx, node)
	}
	di := s.di()
	ch := di.WaitResponse(id)
	if err := di.SendNode(ctx, node); err != nil {
		di.CancelResponse(id, ch)
		return nil, err
	}
	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(15 * time.Second):
		di.CancelResponse(id, ch)
		return nil, nil
	case <-ctx.Done():
		di.CancelResponse(id, ch)
		return nil, ctx.Err()
	}
}

func (s *Socket) GetUSyncDevices(ctx context.Context, jids []types.JID) ([]types.JID, error) {
	return s.cli.GetUserDevices(ctx, jids)
}

func (s *Socket) AssertSessions(ctx context.Context, jids []types.JID, force bool) error {
	return nil
}

func (s *Socket) CreateParticipantNodes(ctx context.Context, devices []types.JID, callKey []byte, encAttrs waBinary.Attrs) ([]waBinary.Node, bool, error) {
	plaintext, err := signaling.EncodeCallKeyMessage(callKey)
	if err != nil {
		return nil, false, err
	}
	id := s.cli.GenerateMessageID()
	return s.di().EncryptMessageForDevices(ctx, devices, id, plaintext, plaintext, encAttrs)
}

func (s *Socket) DecryptCallKey(ctx context.Context, from types.JID, encChild *waBinary.Node) ([]byte, error) {
	typ, _ := encChild.Attrs["type"].(string)
	isPreKey := typ == "pkmsg"
	plaintext, _, err := s.di().DecryptDM(ctx, encChild, from, isPreKey, time.Now())
	if err != nil {
		return nil, err
	}
	return signaling.DecodeCallKeyPlaintext(plaintext)
}

func (s *Socket) GetTCToken(ctx context.Context, jid types.JID) ([]byte, error) {
	if s.cli.Store == nil || s.cli.Store.PrivacyTokens == nil {
		return nil, nil
	}
	resolved, _ := s.ResolveLIDForPN(ctx, jid)
	for _, cand := range []types.JID{resolved.ToNonAD(), jid.ToNonAD()} {
		if cand.IsEmpty() {
			continue
		}
		tok, err := s.cli.Store.PrivacyTokens.GetPrivacyToken(ctx, cand)
		if err != nil {
			return nil, err
		}
		if tok != nil && len(tok.Token) > 0 {
			return tok.Token, nil
		}
	}
	return nil, nil
}

// usableCallLID reports whether a LID can be used as a call destination.
// WhatsApp sometimes stores a placeholder (1099511693312, 2^40+2^16) that
// accepts the offer stanza but never rings the phone. Real account LIDs are
// much larger (around 10^14).
func usableCallLID(lid types.JID) bool {
	if lid.IsEmpty() || lid.Server != types.HiddenUserServer {
		return false
	}
	user, err := strconv.ParseUint(lid.User, 10, 64)
	if err != nil {
		return false
	}
	const minRealLID = 10_000_000_000_000
	return user >= minRealLID
}

// ErrNotOnWhatsApp is returned when WhatsApp reports that none of the number
// spellings we tried are registered.
var ErrNotOnWhatsApp = errors.New("number is not on WhatsApp")

func (s *Socket) ResolveLIDForPN(ctx context.Context, pn types.JID) (types.JID, error) {
	if pn.Server == types.HiddenUserServer {
		return pn, nil
	}
	candidates := []types.JID{pn}
	if alt := brazilMobileAlternate(pn); !alt.IsEmpty() && alt.User != pn.User {
		candidates = append(candidates, alt)
	}
	for _, cand := range candidates {
		if lid := s.lookupLID(ctx, cand); usableCallLID(lid) {
			if cand.User != pn.User {
				slog.Info("resolved LID via alternate brazilian number", "pn", pn.String(), "alt", cand.String(), "lid", lid.String())
			}
			return lid, nil
		}
	}
	// GetUserInfo often returns no LID for the spelling that is not registered.
	// IsOnWhatsApp returns the JID WhatsApp actually has for that phone.
	canon, notRegistered := s.canonicalWhatsAppJID(ctx, candidates)
	if !canon.IsEmpty() && canon.User != pn.User {
		if lid := s.lookupLID(ctx, canon); usableCallLID(lid) {
			slog.Info("resolved LID via whatsapp canonical number", "pn", pn.String(), "canon", canon.String(), "lid", lid.String())
			return lid, nil
		}
		slog.Info("using whatsapp canonical number", "pn", pn.String(), "canon", canon.String())
		return canon, nil
	}
	if notRegistered {
		slog.Warn("number is not on WhatsApp", "pn", pn.String())
		return types.EmptyJID, ErrNotOnWhatsApp
	}
	slog.Warn("no LID for number", "pn", pn.String())
	return pn, nil
}

func (s *Socket) canonicalWhatsAppJID(ctx context.Context, candidates []types.JID) (types.JID, bool) {
	phones := make([]string, 0, len(candidates))
	for _, cand := range candidates {
		if cand.Server == types.DefaultUserServer && cand.User != "" {
			phones = append(phones, "+"+cand.User)
		}
	}
	if len(phones) == 0 {
		return types.EmptyJID, false
	}
	on, err := s.cli.IsOnWhatsApp(ctx, phones)
	if err != nil {
		slog.Warn("is on whatsapp failed", "err", err)
		return types.EmptyJID, false
	}
	var saw bool
	for _, item := range on {
		saw = true
		jid := item.JID.ToNonAD()
		slog.Info("whatsapp number check", "query", item.Query, "jid", jid.String(), "on_whatsapp", item.IsIn)
		if item.IsIn && !jid.IsEmpty() && jid.Server == types.DefaultUserServer {
			return jid, false
		}
	}
	return types.EmptyJID, saw
}

func (s *Socket) lookupLID(ctx context.Context, pn types.JID) types.JID {
	if s.cli.Store != nil && s.cli.Store.LIDs != nil {
		if lid, err := s.cli.Store.LIDs.GetLIDForPN(ctx, pn); err == nil && usableCallLID(lid) {
			return lid
		} else if err == nil && !lid.IsEmpty() {
			slog.Warn("ignoring placeholder LID, refreshing from server", "pn", pn.String(), "lid", lid.String())
		}
	}
	info, err := s.cli.GetUserInfo(ctx, []types.JID{pn})
	if err != nil {
		slog.Warn("get user info failed", "pn", pn.String(), "err", err)
		return types.EmptyJID
	}
	// The server may answer with a different JID than the one we queried
	// (the registered spelling of the number). Any usable LID in the reply counts.
	for jid, ui := range info {
		slog.Info("user info", "query", pn.String(), "jid", jid.String(), "lid", ui.LID.String())
		if usableCallLID(ui.LID) {
			return ui.LID
		}
		if !ui.LID.IsEmpty() {
			slog.Warn("server returned placeholder LID", "pn", pn.String(), "lid", ui.LID.String())
		}
	}
	if len(info) == 0 {
		slog.Info("user info empty", "query", pn.String())
	}
	return types.EmptyJID
}

// brazilMobileAlternate returns the other spelling of a Brazilian mobile:
// 55 + DDD + 9 + 8 digits, or the same number with that 9 removed.
func brazilMobileAlternate(pn types.JID) types.JID {
	if pn.Server != types.DefaultUserServer || !strings.HasPrefix(pn.User, "55") {
		return types.EmptyJID
	}
	u := pn.User
	switch {
	case len(u) == 13 && u[4] == '9':
		return types.NewJID(u[:4]+u[5:], types.DefaultUserServer)
	case len(u) == 12 && u[4] >= '6' && u[4] <= '9':
		return types.NewJID(u[:4]+"9"+u[4:], types.DefaultUserServer)
	default:
		return types.EmptyJID
	}
}
