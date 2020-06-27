Tool uses various method of encryption to ensure maximum privacy (Curve25519, XSalsa20, RSA, Scrypt key derivation function).

Message is encrypted with NaCl Secret Box using https://tweetnacl.js.org/, JavaScript implementation of 
Networking and Cryptography library (NaCl https://nacl.cr.yp.to/). Nonce used for Secret Box is used to generate 
link anchor which is used then to retrieve the message. Nonce is necessary to decrypt the message, it is not 
saved anywhere else so only user using the link can decode the message. To increase security one can use a password. 
This password will be used to generate ephemeral security key. 

Encrypted message with secret key is sealed again using asymmetric algorithm NaCl Box and stored in Database. 

All keys and nonces on browser side are unique for every action. NaCl Box server keys are generated while application 
started for the first time and are encrypted at rest using RSASSA-PSS 3072 bit key with a SHA-256 digest. 
RSA encryption/decryption keys use hardware security module (HSM).

Service runs in Google Cloud RUN infrastructure and is available on https://securenote.io/. 

To Do List:
- [x] internalization support 
- [x] move param link into anchor part
- [x] test the handlers
- [x] password, extra layer of security
- [ ] key encryption key rotation
- [ ] different destruction times
- [ ] more info, faq, etc.
 
Try it locally:
```
go build
./obliviate 
```
and go http://localhost:3000/

For production use you need Google Cloud Run (https://cloud.google.com/run) set and ready and then you need:
- Install GCP SDK (https://cloud.google.com/sdk/docs)
- Firestore DB enabled for your GCP Project (https://cloud.google.com/firestore/docs)
- Google HSM or KMS enabled (https://cloud.google.com/kms/docs)
- Container Registry enabled (https://cloud.google.com/container-registry/docs)
- optional GCP Profiler enabled (https://cloud.google.com/profiler)
- all necessary permission for Firestore, Key and Registry service to use with Cloud Run Service
- modify deploy.sh for your setup
- run docker 
- run ./deploy.sh
- go to Cloud Run Console, create new service, choose deployed container and set environment variables: 
    - ENV = "PROD"
    - HSM_MASTER_KEY = Key version resource ID, see https://cloud.google.com/kms/docs/object-hierarchy#key_version
    - OBLIVIATE_PROJECT_ID = GCP Project ID

If you want to run it without Google Cloud on your private server you should:
- implement your own DataBase interface methods and use it instead of firestore in main.go
- implement your own EncryptionOnRest interface methods if you want to keep the highest security and use it in main.go. 
If not, just mock it.
- get rid of Google Profiler setup from main.go
- set proper listening port in main.go
- if you want to add more languages add them in entries in i18n package
- build and deploy it accordingly to your web server environment demands 

