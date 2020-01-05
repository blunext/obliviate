Tool uses various method of encryption to ensure maximum privacy (Curve25519, XSalsa20, Poly1305, RSA)

Message is ecrypted with NaCl Secret Box using https://tweetnacl.js.org/, JavaScript implementation of 
Networking and Cryptography library (NaCl https://nacl.cr.yp.to/).  Nonce used for Secret Box is used 
as a part of the link generated to retrieve the message. Nonce is necessary to decrypt the message, it is not 
saved anywhere else so only user using the link can decode the message. 

Encrypted message with secret key is sealed again using asymetric algoritm NaCl Box and stored in Database. 

All keys and nonces on browser side are unique for every action. NaCl Box server keys are generated while application 
started for the first time and are encrytped at rest using RSASSA-PSS 3072 bit key with a SHA-256 digest. 
RSA endryption/decryption keys use hardware security module (HSM).

To Do List:
- [x] internalization support 
- [x] move param link into anchor part
- [x] test the handlers
- [ ] password, extra layer of security
- [ ] different destruction times
- [ ] more info, faq, etc.
 