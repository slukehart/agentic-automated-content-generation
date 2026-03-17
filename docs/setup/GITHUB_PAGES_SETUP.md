# GitHub Pages Setup Guide

This guide helps you enable GitHub Pages for your TikTok OAuth callback and legal documents.

---

## Why GitHub Pages?

**TikTok requires:**
- ✅ Public HTTPS URLs for Privacy Policy and Terms of Service
- ✅ Public callback URL for OAuth (doesn't accept localhost)

**GitHub Pages provides:**
- Free HTTPS hosting
- Simple setup (no coding required)
- Automatic deployment from your repository

---

## Quick Setup (5 minutes)

### Step 1: Push Files to GitHub

Make sure these files are in your repository:
- `callback.html` - OAuth callback page
- `PRIVACY_POLICY.md` - Privacy policy
- `TERMS_OF_SERVICE.md` - Terms of service

```bash
git add callback.html PRIVACY_POLICY.md TERMS_OF_SERVICE.md
git commit -m "Add TikTok OAuth callback and legal docs"
git push origin main
```

### Step 2: Enable GitHub Pages

1. Go to your repository: `https://github.com/slukehart/content-generation-automation`
2. Click **Settings** (top navigation)
3. Scroll down and click **Pages** (left sidebar)
4. Under **Source**:
   - Branch: Select `main` (or `master`)
   - Folder: Select `/ (root)`
5. Click **Save**
6. Wait 1-2 minutes for deployment

### Step 3: Verify Your URLs

After deployment, test these URLs in your browser:

**OAuth Callback:**
```
https://slukehart.github.io/content-generation-automation/callback.html
```
Should show: "Processing authorization..." message

**Privacy Policy:**
```
https://slukehart.github.io/content-generation-automation/PRIVACY_POLICY
```
Should show: Your privacy policy (formatted as HTML)

**Terms of Service:**
```
https://slukehart.github.io/content-generation-automation/TERMS_OF_SERVICE
```
Should show: Your terms of service (formatted as HTML)

### Step 4: Configure TikTok Developer Portal

1. Go to [TikTok Developer Portal](https://developers.tiktok.com/apps)
2. Select your app
3. Go to **App Settings**
4. Add **Redirect URI:**
   ```
   https://slukehart.github.io/content-generation-automation/callback.html
   ```
5. Add **Privacy Policy URL:**
   ```
   https://slukehart.github.io/content-generation-automation/PRIVACY_POLICY
   ```
6. Add **Terms of Service URL:**
   ```
   https://slukehart.github.io/content-generation-automation/TERMS_OF_SERVICE
   ```
7. **Save** all changes

---

## Testing the OAuth Flow

After setup:

1. Run your app:
   ```bash
   go run main.go -upload-tiktok
   ```

2. Open the authorization URL in your browser

3. After authorizing, you'll be redirected to:
   ```
   https://slukehart.github.io/content-generation-automation/callback.html?code=ABC123...
   ```

4. The page will display your code beautifully:
   - Shows the authorization code
   - Has a "Copy Code" button
   - Provides instructions

5. Click "Copy Code" and paste into your terminal

---

## Troubleshooting

### GitHub Pages not working?

**Check deployment status:**
1. Go to repository **Actions** tab
2. Look for "pages build and deployment"
3. Wait for green checkmark

**Try these URLs:**
- With trailing slash: `https://slukehart.github.io/content-generation-automation/`
- Direct file: `https://slukehart.github.io/content-generation-automation/callback.html`

### TikTok still rejecting callback?

**Verify:**
- [ ] Redirect URI in TikTok matches exactly (including `.html`)
- [ ] GitHub Pages is deployed (check the URLs work)
- [ ] No typos in the URLs
- [ ] Using HTTPS (not HTTP)

### OAuth error: "redirect_uri_mismatch"

**Cause:** The redirect URI in your code doesn't match TikTok's registered URI

**Fix:** Update `media/tiktok_auth.go` line with the correct URL:
```go
tiktokRedirectURI = "https://slukehart.github.io/content-generation-automation/callback.html"
```

---

## Custom Domain (Optional)

Want to use your own domain like `https://yourdomain.com/callback`?

1. Add a `CNAME` file to your repository:
   ```
   yourdomain.com
   ```

2. Configure DNS:
   - Add CNAME record pointing to `slukehart.github.io`

3. Update TikTok redirect URI to your custom domain

4. Update Go code with your custom domain

---

## Security Notes

**GitHub Pages is public:**
- ✅ Safe for callback pages (no sensitive data)
- ✅ Safe for legal documents (meant to be public)
- ❌ Never put credentials or secrets in repository

**Callback page security:**
- Authorization code is only valid for 60 seconds
- Code is single-use (cannot be reused)
- Code is shown client-side (not logged by GitHub)
- User must paste code into their terminal immediately

---

## Alternative: Use Your Own Domain

If you have your own website, you can host these files there instead:

1. Upload `callback.html`, privacy policy, and terms
2. Use your domain URLs in TikTok Developer Portal
3. Update `media/tiktok_auth.go` with your URLs

**Example:**
```go
tiktokRedirectURI = "https://yourdomain.com/callback.html"
```

---

## For TikTok App Review

When submitting your app, provide:

**Redirect URI:**
```
https://slukehart.github.io/content-generation-automation/callback.html
```

**Privacy Policy:**
```
https://slukehart.github.io/content-generation-automation/PRIVACY_POLICY
```

**Terms of Service:**
```
https://slukehart.github.io/content-generation-automation/TERMS_OF_SERVICE
```

**Demo Video:** Show the OAuth flow working with these URLs

---

## Need Help?

- **GitHub Pages docs:** https://docs.github.com/en/pages
- **TikTok OAuth docs:** https://developers.tiktok.com/doc/oauth
- **Issues?** Check repository Issues tab

---

**Last Updated:** January 22, 2026
