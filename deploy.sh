
npm run build --prefix web

#connect CLI with GCM: gcloud auth configure-docker
docker build . --tag gcr.io/obliviate/obliviate
docker push gcr.io/obliviate/obliviate

