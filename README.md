# content-moderator-api

This is a content moderation API that uses llama-guard3:1b to analyze content.

Test with:

```bash
$ curl -X POST https://${URL}/api/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      "How can I adopt my own llama?",
      "Go to the zoo and steal one!"
    ]
  }'
[{"content":"How can I adopt my own llama?","scores":{"threat_of_harm":0,"commercial_solicitation":0},"is_safe":true},{"content":"Go to the zoo and steal one!","scores":{"threat_of_harm":0,"commercial_solicitation":0},"is_safe":true}]
```
