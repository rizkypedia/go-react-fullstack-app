import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import App from '../../src/components/App';

function jsonResponse(body: object): Response {
  return {
    ok: true,
    status: 200,
    json: async () => body,
  } as Response;
}

describe('App', () => {
  afterEach(() => {
    Reflect.deleteProperty(globalThis, 'fetch');
  });

  it('authenticates and sends the input value to the protected API', async () => {
    const fetchMock = jest.fn()
      .mockResolvedValueOnce(
        jsonResponse({ access_token: 'test-token', token_type: 'Bearer' }),
      )
      .mockResolvedValueOnce(
        jsonResponse({ value: 'hello from React' }),
      );
    Object.defineProperty(globalThis, 'fetch', { configurable: true, value: fetchMock });
    render(<App />);

    fireEvent.change(screen.getByRole('textbox', { name: /username/i }), {
      target: { value: 'demo' },
    });
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'secret' } });
    fireEvent.click(screen.getByRole('button', { name: 'Sign in' }));
    await screen.findByRole('heading', { name: 'Protected API request' });

    fireEvent.change(screen.getByRole('textbox', { name: /value to send/i }), {
      target: { value: 'hello from React' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Send to API' }));

    expect(await screen.findByRole('status')).toHaveTextContent('hello from React');
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      '/api/echo',
      expect.objectContaining({
        headers: expect.objectContaining({ Authorization: 'Bearer test-token' }),
        body: JSON.stringify({ value: 'hello from React' }),
      }),
    );
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
  });

  it('revokes the access token on sign out', async () => {
    const fetchMock = jest.fn()
      .mockResolvedValueOnce(jsonResponse({ access_token: 'test-token', token_type: 'Bearer' }))
      .mockResolvedValueOnce({ ok: true, status: 204 });
    Object.defineProperty(globalThis, 'fetch', { configurable: true, value: fetchMock });
    render(<App />);

    fireEvent.change(screen.getByRole('textbox', { name: /username/i }), {
      target: { value: 'demo' },
    });
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'secret' } });
    fireEvent.click(screen.getByRole('button', { name: 'Sign in' }));
    await screen.findByRole('heading', { name: 'Protected API request' });

    fireEvent.click(screen.getByRole('button', { name: 'Sign out' }));
    await screen.findByRole('heading', { name: 'Development sign in' });

    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      '/api/auth/logout',
      expect.objectContaining({
        method: 'POST',
        headers: { Authorization: 'Bearer test-token' },
      }),
    );
  });
});
