import { useState } from "react";

import { login, signup } from "../../api/authApi";

export default function AuthPage({
  page,
  onPageChange,
  onSessionReady,
  onMessage,
  uiPreview,
}) {
  const [loading, setLoading] = useState(false);
  const [loginForm, setLoginForm] = useState({ username: "", password: "", remember: true });
  const [signupForm, setSignupForm] = useState({ username: "", email: "", password: "" });

  async function handleLogin(event) {
    event.preventDefault();
    if (uiPreview) {
      onSessionReady("preview-token", "preview-refresh-token", loginForm.remember);
      onMessage({ type: "success", text: "Preview mode: открыт экран заметок." });
      return;
    }

    try {
      setLoading(true);
      const authData = await login(loginForm.username, loginForm.password);
      onSessionReady(authData.token, authData.refresh_token, loginForm.remember);
      onMessage({ type: "success", text: "Вход выполнен." });
    } catch (error) {
      onMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Произошла ошибка.",
      });
    } finally {
      setLoading(false);
    }
  }

  async function handleSignup(event) {
    event.preventDefault();
    if (uiPreview) {
      onSessionReady("preview-token", "preview-refresh-token", true);
      onMessage({ type: "success", text: "Preview mode: открыт экран заметок после регистрации." });
      return;
    }

    try {
      setLoading(true);
      const authData = await signup(signupForm);
      onSessionReady(authData.token, authData.refresh_token, true);
      onMessage({ type: "success", text: "Аккаунт создан." });
    } catch (error) {
      onMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Произошла ошибка.",
      });
    } finally {
      setLoading(false);
    }
  }

  if (page === "signup") {
    return (
      <AuthFormLayout
        title="Регистрация"
        actionLabel="Создать аккаунт"
        footerLabel="Уже есть аккаунт?"
        footerAction="Вход"
        onFooterClick={() => onPageChange("login")}
        onSubmit={handleSignup}
        loading={loading}
      >
        <LabeledInput
          label="Username"
          value={signupForm.username}
          placeholder="Придумай логин"
          onChange={(value) => setSignupForm((current) => ({ ...current, username: value }))}
        />
        <LabeledInput
          label="Email"
          type="email"
          value={signupForm.email}
          placeholder="mail@example.com"
          onChange={(value) => setSignupForm((current) => ({ ...current, email: value }))}
        />
        <LabeledInput
          label="Пароль"
          type="password"
          value={signupForm.password}
          placeholder="Не менее 6 символов"
          onChange={(value) => setSignupForm((current) => ({ ...current, password: value }))}
        />
      </AuthFormLayout>
    );
  }

  return (
    <AuthFormLayout
      title="Вход"
      actionLabel="Войти"
      footerLabel="Нет аккаунта?"
      footerAction="Регистрация"
      onFooterClick={() => onPageChange("signup")}
      onSubmit={handleLogin}
      loading={loading}
    >
      <LabeledInput
        label="Email / Username"
        value={loginForm.username}
        placeholder="Введите логин"
        onChange={(value) => setLoginForm((current) => ({ ...current, username: value }))}
      />
      <LabeledInput
        label="Пароль"
        type="password"
        value={loginForm.password}
        placeholder="Password"
        onChange={(value) => setLoginForm((current) => ({ ...current, password: value }))}
      />
      <label className="remember-row">
        <input
          type="checkbox"
          checked={loginForm.remember}
          onChange={(event) => setLoginForm((current) => ({ ...current, remember: event.target.checked }))}
        />
        <span>Запомнить меня</span>
      </label>
    </AuthFormLayout>
  );
}

function AuthFormLayout({
  title,
  actionLabel,
  footerLabel,
  footerAction,
  onFooterClick,
  onSubmit,
  loading,
  children,
}) {
  return (
    <main className="auth-page">
      <section className="auth-panel">
        <header className="auth-header">
          <h1>{title}</h1>
        </header>
        <form className="auth-form" onSubmit={onSubmit}>
          {children}
          <button className="primary-button" type="submit" disabled={loading}>
            {actionLabel}
          </button>
        </form>
        <div className="auth-footer">
          <span>{footerLabel}</span>
          <button className="text-button" onClick={onFooterClick} type="button">
            {footerAction}
          </button>
        </div>
      </section>
    </main>
  );
}

function LabeledInput({ label, value, onChange, placeholder, type = "text" }) {
  return (
    <label className="field-row">
      <span>{label}</span>
      <input type={type} value={value} onChange={(event) => onChange(event.target.value)} placeholder={placeholder} />
    </label>
  );
}
