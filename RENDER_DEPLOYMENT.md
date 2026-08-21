# Deploy ezBookkeeping on Render (Free)

This repository is set up to deploy the latest source on Render and to check the
upstream project for updates every day.

## Important limits of the Free plan

- A free Render web service can stop after 15 minutes without visitors. The next
  visit starts it again, so the first page load can be slow.
- Free web services do not have a persistent disk. This deployment uses
  PostgreSQL for bookkeeping data and disables custom icons, avatars stored by
  the app, and receipt picture uploads so those files are not lost on restart.
- Render's free PostgreSQL option is temporary. Before it expires, migrate to a
  paid Render PostgreSQL plan or set `EBK_DATABASE_URL` to a managed external
  PostgreSQL database. Export your data before making any migration.

## First deployment

1. Create an empty GitHub repository in your own account, then push this folder
   to its `main` branch. Do not deploy straight from the original upstream
   repository, because it does not contain your configuration and changes.
2. In the GitHub repository, open **Settings → Actions → General**, set
   **Workflow permissions** to **Read and write permissions**, then save. This
   permits the updater to push upstream changes back to your repository.
3. In Render, select **New → Blueprint**, connect that GitHub repository and
   approve the `render.yaml` configuration. It creates a Free web service and a
   Free PostgreSQL database in Singapore.
4. Wait for the first Docker build to finish. Open the generated
   `https://…onrender.com` address and create your user account.
5. For a personal-only finance site, set the Render environment variable
   `EBK_USER_ENABLE_REGISTER` to `false` after creating your account, then
   redeploy. Do not expose the registration page to other people unnecessarily.
6. To use the member administration page, set the three Render secrets
   `EBK_SECURITY_ROOT_ADMIN_USERNAME`, `EBK_SECURITY_ROOT_ADMIN_EMAIL` and
   `EBK_SECURITY_ROOT_ADMIN_PASSWORD`. On the next deployment, the root account
   is created once and protected from removal. Open `https://…onrender.com/admin`
   with that account. The root account can grant or remove administrator access;
   administrators can manage other member accounts and data.

## Updating source code

There are two options:

- **Automatic:** every day at approximately 09:17 Vietnam time, GitHub Actions
  fetches and merges the newest `main` branch from
  `mayswind/ezbookkeeping`. A successful merge is pushed to your `main` branch;
  Render then builds and deploys it automatically.
- **Update now:** open the **Actions** tab in your GitHub repository, select
  **Update from ezBookkeeping upstream**, click **Run workflow**, then choose
  **Run workflow** again. This is the manual update button.

The updater never force-pushes. If your custom code conflicts with an upstream
change, it stops without deploying anything. Resolve the conflict in GitHub or
locally, then run the workflow again.

## Re-enabling picture uploads

Picture files cannot safely live on a Free Render web service. To use receipt
images or custom avatars, configure WebDAV or MinIO/S3-compatible object
storage first, then update the `storage` environment settings and enable the
corresponding user features. Keep `EBK_DATABASE_URL` as a secret in Render; do
not commit database credentials to GitHub.
