import { useState, type SubmitEvent } from 'react';
import {
  Alert,
  Box,
  Button,
  Container,
  Paper,
  Stack,
  TextField,
  Typography,
} from '@mui/material';

type LoginResponse = {
  access_token: string;
};

type EchoResponse = {
  value: string;
};

async function readError(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as { error?: string };
    return body.error ?? `Request failed (${response.status})`;
  } catch {
    return `Request failed (${response.status})`;
  }
}

export default function App() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [token, setToken] = useState('');
  const [value, setValue] = useState('');
  const [result, setResult] = useState('');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  async function login(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    setError('');
    setBusy(true);
    try {
      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });
      if (!response.ok) {
        throw new Error(await readError(response));
      }
      const body = (await response.json()) as LoginResponse;
      setToken(body.access_token);
      setPassword('');
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Could not sign in');
    } finally {
      setBusy(false);
    }
  }

  async function submitValue(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    setError('');
    setResult('');
    setBusy(true);
    try {
      const response = await fetch('/api/echo', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ value }),
      });
      if (!response.ok) {
        throw new Error(await readError(response));
      }
      const body = (await response.json()) as EchoResponse;
      setResult(body.value);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Could not send the value');
    } finally {
      setBusy(false);
    }
  }

  async function signOut() {
    setBusy(true);
    setError('');
    try {
      const response = await fetch('/api/auth/logout', {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!response.ok) {
        throw new Error(await readError(response));
      }
    } catch (requestError) {
      const message = requestError instanceof Error ? requestError.message : 'Could not sign out';
      setError(`${message}. The token may remain valid until it expires.`);
    } finally {
      setToken('');
      setResult('');
      setBusy(false);
    }
  }

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'grid',
        placeItems: 'center',
        px: 2,
        py: 6,
        background:
          'radial-gradient(circle at 12% 12%, rgba(77, 112, 230, 0.12), transparent 32%), #f5f7fa',
      }}
    >
      <Container maxWidth="sm">
        <Paper elevation={0} sx={{ p: { xs: 3, sm: 5 }, border: '1px solid', borderColor: 'divider' }}>
          <Stack spacing={3}>
            <Box>
              <Typography color="primary" fontWeight={700} letterSpacing={1.5} variant="overline">
                React + Go starter
              </Typography>
              <Typography component="h1" variant="h4" fontWeight={750} sx={{ mt: 0.5 }}>
                Send a message
              </Typography>
              <Typography color="text.secondary" sx={{ mt: 1 }}>
                Sign in to receive a 12-hour token, then send a value to the protected Go API.
              </Typography>
            </Box>

            {error && <Alert severity="error">{error}</Alert>}

            {!token ? (
              <Stack component="form" onSubmit={login} spacing={2}>
                <Typography component="h2" variant="h6">Development sign in</Typography>
                <TextField
                  autoComplete="username"
                  label="Username"
                  onChange={(event) => setUsername(event.target.value)}
                  required
                  value={username}
                />
                <TextField
                  autoComplete="current-password"
                  label="Password"
                  onChange={(event) => setPassword(event.target.value)}
                  required
                  type="password"
                  value={password}
                />
                <Button disabled={busy} type="submit" variant="contained" size="large">
                  {busy ? 'Signing in…' : 'Sign in'}
                </Button>
              </Stack>
            ) : (
              <Stack component="form" onSubmit={submitValue} spacing={2}>
                <Typography component="h2" variant="h6">Protected API request</Typography>
                <TextField
                  autoFocus
                  label="Value to send"
                  onChange={(event) => setValue(event.target.value)}
                  required
                  value={value}
                />
                <Stack direction="row" spacing={1.5}>
                  <Button disabled={busy} type="submit" variant="contained" size="large">
                    {busy ? 'Sending…' : 'Send to API'}
                  </Button>
                  <Button
                    disabled={busy}
                    onClick={signOut}
                    size="large"
                    type="button"
                  >
                    {busy ? 'Signing out…' : 'Sign out'}
                  </Button>
                </Stack>
                {result && (
                  <Alert severity="success" role="status">
                    API response: <strong>{result}</strong>
                  </Alert>
                )}
              </Stack>
            )}
          </Stack>
        </Paper>
      </Container>
    </Box>
  );
}
