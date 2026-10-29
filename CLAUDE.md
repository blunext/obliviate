# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Obliviate is a secure self-destructing notes service (securenote.io). Messages are encrypted client-side using NaCl (TweetNaCl.js) and stored encrypted in Firestore. Server keys are encrypted at rest using Google Cloud KMS.

## Build & Run Commands

```bash
# Local development (uses mock DB and encryption)
npm install --prefix web
npm run build --prefix web
go run .
# Visit http://localhost:3000/

# Run all tests
go test ./...

# Run single test
go test -run TestName ./path/to/package

# Frontend dev server (with hot reload, http://localhost:5173/; needs `go run .` on :3000 for the API)
cd web && npm run dev

# Type check frontend
cd web && npm run check
```

The Go binary embeds `web/build` at compile time (`//go:embed`), so after a frontend change run
`npm run build --prefix web` and restart `go run .`.

### GUI end-to-end test

`web/tests/e2e.js` is a script for the Playwright MCP tool `browser_run_code_unsafe` (no npm
dependency). After frontend changes: `npm run build --prefix web`, start `go run .` in the
background, run the tool with `filename: "web/tests/e2e.js"`, then stop the server and remove
`.playwright-mcp/`. It reports PASS/FAIL for encrypt/decrypt with and without password, the
message-read screen, the read limit, corrupted links and network errors. Extend it when adding UI
behavior.

## Architecture

### Encryption Flow
1. **Client**: Message encrypted with NaCl SecretBox. The key is random, or derived from the
   password with scrypt (only the salt is sent then). The SecretBox nonce goes into the link
   (`/?<part>#<rest>`) and never reaches the server; `nacl.hash(nonce)` is the message ID.
2. **Client→Server**: Key (or salt) + ciphertext sealed with NaCl Box for the server's public key.
3. **Storage**: The Box is stored as-is in Firestore; messages expire after 4 weeks.
4. **Read**: Server opens the Box and re-seals it for the reader's ephemeral public key
   (`ProcessRead`); the client decrypts it and the SecretBox.
5. **Server keys**: NaCl keypair encrypted at rest with Cloud KMS (RSA, `HSM_MASTER_KEY`).

### HTTP API (`main.go`, stdlib `http.ServeMux`)

| Route | Purpose |
|-------|---------|
| `GET /variables` | Localized UI texts + server public key (frontend loads it in `+page.js`) |
| `POST /save` | Store a message |
| `POST /read` | Read a message (404 when read, expired or unknown) |
| `DELETE /delete` | Delete a password-protected message after successful decryption |
| `DELETE /expired` | Remove expired messages |
| `GET /` | Embedded static frontend |

### Key Packages

| Package | Purpose |
|---------|---------|
| `app/` | Business logic (ProcessSave, ProcessRead, ProcessDelete, ProcessDeleteExpired) |
| `handler/` | HTTP handlers; `handler/webModels/` request/response types |
| `repository/` | Firestore operations and sharded message counter; `mock/` for local dev, `model/` data types |
| `crypt/` | NaCl key management |
| `crypt/rsa/` | Cloud KMS integration |
| `config/` | Runtime configuration |
| `logs/` | slog handler for Cloud Logging |
| `i18n/` | Internationalization; all UI texts live in `entries.go`, not in the frontend |
| `web/` | SvelteKit frontend (adapter-static, Bootstrap styling) |

### Environment Modes
- `ENV=PROD`: Uses Firestore + Cloud KMS (`STAGE` other than `prod` uses `test_` collections)
- Local (default): Uses `mock.StorageMock()` + `rsa.NewMockAlgorithm()`, in-memory data
- Other variables: `PORT` (default 3000), `HSM_MASTER_KEY`, `KMS_CREDENTIAL_FILE`,
  `FIRESTORE_CREDENTIAL_FILE`, `OBLIVIATE_PROJECT_ID`

### Password-Protected Messages
- Link query part has 4 characters (3 without password); client sends `password: true` in request
- Server uses `GetMessageWithReadLimit()` - max 10 reads then auto-delete
- Decryption happens client-side (scrypt + nacl); after success the client calls `DELETE /delete`

## Frontend Structure

- `web/src/routes/+page.js` - Loads `/variables` (texts + server public key)
- `web/src/routes/+page.svelte` - Main page, switches between the views below
- `web/src/routes/Encrypt.svelte` - Message creation form
- `web/src/routes/ShowLink.svelte` - Generated link with copy button
- `web/src/routes/Decrypt.svelte` - Message reading (and message-read info on 404)
- `web/src/routes/ShowMessage.svelte` - Decrypted message
- `web/src/routes/Commons.js` - API URLs, `post()` helper, scrypt key derivation
- `web/src/lib/` - Reusable components (ShieldIcon, EyeIcon)
- Uses zxcvbn-ts for password strength
