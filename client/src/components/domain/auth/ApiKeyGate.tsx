import { useEffect, useState, type FormEvent, type ReactNode } from "react";
import { ApiError, apiGet } from "@/lib/api";
import { onSignOut, setApiKey } from "@/lib/api-key";
import { resetSessionsWire } from "@/stores/sessions";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export const ApiKeyGate = ({ children }: { children: ReactNode }) => {
  const [state, setState] = useState<"checking" | "ok" | "locked">("checking");
  const [draft, setDraft] = useState("");
  const [error, setError] = useState("");

  const probe = () => {
    setState("checking");
    apiGet("/api/sessions")
      .then(() => setState("ok"))
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 401) {
          setApiKey("");
          setState("locked");
          return;
        }
        setState("ok");
      });
  };

  useEffect(() => {
    probe();
    return onSignOut(() => {
      resetSessionsWire();
      setDraft("");
      setError("");
      setState("locked");
    });
  }, []);

  const submit = (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setApiKey(draft);
    apiGet("/api/sessions")
      .then(() => setState("ok"))
      .catch((err: unknown) => {
        setApiKey("");
        setError(err instanceof ApiError && err.status === 401 ? "Chave inválida" : "Não foi possível conectar");
        setState("locked");
      });
  };

  if (state === "checking") {
    return <div className="grid min-h-screen place-items-center text-sm text-muted-foreground">Conectando…</div>;
  }

  if (state === "locked") {
    return (
      <div className="grid min-h-screen place-items-center p-6">
        <form onSubmit={submit} className="w-full max-w-sm space-y-4">
          <div>
            <h1 className="text-lg font-semibold">Chave da API</h1>
            <p className="text-sm text-muted-foreground">Informe a chave de acesso da API.</p>
          </div>
          <Input
            type="password"
            autoFocus
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            placeholder="Chave"
            autoComplete="current-password"
          />
          {error ? <p className="text-sm text-destructive">{error}</p> : null}
          <Button type="submit" className="w-full" disabled={!draft.trim()}>
            Entrar
          </Button>
        </form>
      </div>
    );
  }

  return children;
};
