---
title: "Data & Export"
description: "Export your collection and manage your JukeLab data"
order: 3
---

## Data Export

**Available on: Connoisseur plan**

Your collection data belongs to you. Connoisseur users can export their complete album list at any time.

## What You Can Export

### Complete Collection Data
- Album titles and artists
- Track listings
- Spotify links (where applicable)
- Jukebox settings

### Export Format

Data is exported as **JSON** format, making it:
- Easy to import into spreadsheets
- Compatible with other tools
- Human-readable
- Great for backups

## How to Export

### Export Your Collection

1. Go to your **Jukebox Settings**
2. Click **"Export Collection"**
3. Download your data file

That's it! Your complete collection is now saved locally.

### What's in the Export

```json
{
  "jukebox": {
    "name": "My House Party",
    "created": "2024-01-15"
  },
  "albums": [
    {
      "title": "Rumours",
      "artist": "Fleetwood Mac",
      "spotify_id": "...",
      "tracks": 11
    }
  ],
  "total_albums": 100,
  "total_tracks": 1247
}
```

## Use Cases

### Backup

Keep a backup of your carefully curated collection. If you ever need to recreate your jukebox, you have the data.

### Sharing

Share your collection list with friends who want inspiration for their own jukeboxes.

### Migration

Move your collection to a different jukebox or share album ideas with other JukeLab users.

### Analysis

See stats about your collection:
- Total albums and tracks
- Most represented artists
- Genre distribution

## Upgrade to Export

Export is a Connoisseur feature. Nerd plan users can upgrade anytime.

### Why Connoisseur Only?

Data export is included in the Connoisseur plan alongside other premium features like offline mode and MP3 support.

Nerd users still own their data and can view it in-app. Export makes it portable.

→ [Compare Plans](/docs/plans/features)

## Privacy & Security

### Your Data is Yours

- We never sell your data
- We never share it with third parties
- Export lets you take it anywhere

### What We Don't Track

- What songs get played at your parties
- Your guests' information
- Listening habits or analytics
- Anything we could sell

## Deleting Data

### Delete a Jukebox

1. Go to jukebox settings
2. Click **"Delete Jukebox"**
3. Confirm deletion

This removes the jukebox and collection. The guest link becomes invalid.

### Delete Your Account

1. Go to Account Settings
2. Click **"Delete Account"**
3. Confirm deletion

This permanently removes all your data. Export first if you want to keep it.

## Next Steps

- [Collection Management](/docs/features/collection-management) - Organize your albums
- [Upgrade to Connoisseur](/app/billing) - Get export access
- [Plans & Features](/docs/plans/features) - See all features
