# Operations

## Data persistence on Firebase App Hosting

App Hosting runs on Cloud Run, whose disk is ephemeral: anything under `DATA_DIR` (`/tmp/chorus-repository`) disappears when the instance restarts or is replaced. The App Hosting runtime image also has no `git` binary, so the hosted app stores plain JSON files and makes no Git commits (the Docker image does include git).

To keep the data, the server keeps a snapshot of `DATA_DIR` in a private Cloud Storage bucket (`internal/backup`):

- **Start-up:** if `DATA_DIR` is empty, the snapshot is downloaded and unpacked. If the download fails for any reason other than "no snapshot yet", the server **exits** instead of starting empty, because serving an empty directory would let the next backup replace the stored data.
- **After writes:** a successful `POST`/`PUT`/`PATCH`/`DELETE` is answered only after the snapshot is uploaded (unchanged data is skipped). A failed upload is logged and retried; it does not fail the request.
- **Consistency:** a snapshot can, rarely, catch a file while a concurrent request is rewriting it. The writer's own request flushes again right afterwards, which replaces that snapshot, so only a crash inside that window could keep it. Making file writes atomic is planned together with confining the store to its data root.
- **Periodically and on shutdown:** a 60 s safety-net flush, and a final flush on `SIGTERM`.
- **One writer:** uploads use `ifGenerationMatch`, so a second instance cannot overwrite a newer snapshot (it logs a conflict instead). `apphosting.yaml` therefore sets `maxInstances: 1`. During a rollout the old and new instance briefly overlap; writes the old instance accepts after the new one has restored can be lost.

Backups are on when `CHORUS_BACKUP_BUCKET` is set (optionally `CHORUS_BACKUP_OBJECT`, default `chorus-data.tar.gz`). The instance's service account must be able to read and write objects in the bucket.

### One-time setup

```bash
PROJECT=chorus-chat-app
BUCKET=chorus-chat-app-chorus-data
SA=firebase-app-hosting-compute@${PROJECT}.iam.gserviceaccount.com

gcloud storage buckets create "gs://${BUCKET}" --project "$PROJECT" --location europe-west4 \
  --uniform-bucket-level-access --public-access-prevention
gcloud storage buckets update "gs://${BUCKET}" --versioning        # keep older snapshots
gcloud storage buckets add-iam-policy-binding "gs://${BUCKET}" \
  --member "serviceAccount:${SA}" --role roles/storage.objectUser
```

Create the bucket and grant access **before** merging a change that sets `CHORUS_BACKUP_BUCKET`, otherwise the new revision refuses to start and the rollout fails (the previous revision keeps serving).

### Restoring an older snapshot

Versioning keeps previous generations of the object:

```bash
gcloud storage ls --all-versions "gs://${BUCKET}/chorus-data.tar.gz"
gcloud storage cp "gs://${BUCKET}/chorus-data.tar.gz#<generation>" "gs://${BUCKET}/chorus-data.tar.gz"
```

Then start a new revision so the instance restores from it.

### Limits

The snapshot is built in memory and uploaded in one request, so it suits a small data set (a warning is logged above 100 MB). When the data outgrows that, move `DATA_DIR` to a persistent disk (a VM or a Cloud Run volume).
