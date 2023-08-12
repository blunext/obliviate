Tool uses various method of encryption to ensure maximum privacy (Curve25519, XSalsa20, RSA, Scrypt key derivation function).

Message is encrypted with NaCl Secret Box using https://tweetnacl.js.org/, JavaScript implementation of
Networking and Cryptography library (NaCl https://nacl.cr.yp.to/). Nonce used for Secret Box is used to generate
link anchor which is used then to retrieve the message. Nonce is necessary to decrypt the message, it is not
saved anywhere else so only user using the link can decode the message. To increase security one can use a password.
This password will be used to generate ephemeral security key.

Encrypted message with secret key is sealed again using asymmetric algorithm NaCl Box and stored in Database.

All keys and nonces on browser side are unique for every action. NaCl Box server keys are generated while application
started for the first time and are encrypted at rest using RSASSA-PSS 3072 bit key with a SHA-256 digest.
RSA encryption/decryption keys use Cloud KMS module.

Service runs in Google Cloud RUN infrastructure and is available on https://securenote.io/.

Try it locally:
```
npm install --prefix web
npm run build --prefix web
go run .

and go http://localhost:3000/
```
---
This project is tested with BrowserStack.