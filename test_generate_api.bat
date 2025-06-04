@echo off
echo Testing Generate Scheduled Item API...

echo.
echo 1. Test generating a scheduled item from prompt
curl -X POST http://localhost:8080/generate-scheduled-item -H "Content-Type: application/json" -d "{\"prompt\":\"Schedule a team meeting for next Monday at 2 PM\"}"

echo.
echo 2. Test with empty prompt (should fail)
curl -X POST http://localhost:8080/generate-scheduled-item -H "Content-Type: application/json" -d "{\"prompt\":\"\"}"

echo.
echo 3. Test with invalid JSON (should fail)
curl -X POST http://localhost:8080/generate-scheduled-item -H "Content-Type: application/json" -d "invalid json"

echo.
echo 4. Test with complex recurring task
curl -X POST http://localhost:8080/generate-scheduled-item -H "Content-Type: application/json" -d "{\"prompt\":\"Set up a weekly standup meeting every Monday at 9 AM starting next week, expires end of year\"}"

echo.
echo Testing completed!