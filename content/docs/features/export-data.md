---
title: "Export Data"
description: "Export your goals and progress data (Pro & Enterprise feature)"
order: 3
---

## Data Export

**Available on: Pro and Enterprise plans**

Your progress data belongs to you. Pro and Enterprise users can export their complete goal history at any time.

## What You Can Export

### Complete Goal Data
- Goal titles and descriptions
- All logged repetitions with dates
- Progress percentages
- Completion status
- Created and updated timestamps

### Export Format

Data is exported as **JSON** format, making it:
- Easy to import into spreadsheets (Excel, Google Sheets)
- Compatible with data analysis tools
- Human-readable
- Machine-parseable for custom tools

## How to Export

### Export All Goals

1. Go to your **Goals** page
2. Click **"Export Goals"** button
3. Download `goals-export.json`

That's it! Your complete data is now saved locally.

### What's in the Export

```json
{
  "exported_at": "2024-10-29T12:00:00Z",
  "goals": [
    {
      "id": "goal_abc123",
      "title": "Practice Guitar Daily",
      "description": "20 minutes of focused practice",
      "current_step": 47,
      "status": "active",
      "created_at": "2024-09-12T08:00:00Z",
      "updated_at": "2024-10-29T11:30:00Z"
    }
  ]
}
```

## Use Cases

### Personal Analysis
Import into a spreadsheet to analyze:
- Your completion rate
- Average time to reach 100
- Most successful goal types
- Patterns in your progress

### Backup
Keep a local backup of your progress. While we backup all data, some users prefer having their own copy.

### Portfolio/Resume
Showcase your completed goals and consistency:
- "Coded for 100 consecutive days"
- "Completed 5 different 100-day challenges"
- "Maintained 3 simultaneous goals for 100 days"

### Custom Tools
Build custom tools or integrations:
- Import into other tracking apps
- Create custom visualizations
- Analyze with data science tools
- Build personal dashboards

## Upgrade to Export

Export is a Pro/Enterprise feature. Free plan users can upgrade anytime.

### Why Pro/Enterprise Only?

Data export requires additional infrastructure and bandwidth. By limiting it to paid plans, we can:
- Provide reliable, fast exports
- Maintain data integrity
- Support continued development

Free users still own their data and can always view it in-app. Export just makes it portable.

→ [Compare Plans](/docs/plans/features)

## Privacy & Security

### Your Data is Yours
- We never sell your data
- We never share it with third parties
- Export lets you take it anywhere

### What We See
When you export data, we log:
- That an export was requested (for security)
- Timestamp

We don't see:
- What you do with exported data
- Where you store it
- How you use it

## API Access (Enterprise)

Enterprise users can access their data via API:
- Programmatic data access
- Real-time sync
- Build custom integrations
- Automate workflows

Contact us for API documentation and credentials.

→ [Enterprise Plans](/docs/plans/features)

## Next Steps

- [Goal Management](/docs/features/goal-management) - Organize your goals
- [Upgrade to Pro](/app/billing) - Get export access
- [Plans & Features](/docs/plans/features) - See all features
