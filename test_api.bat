@echo off
echo Testing REST API...

echo.
echo 1. Get all scheduled items (should show initial sample data)
curl -X GET http://localhost:8080/scheduled-items

echo.
echo 2. Create a new scheduled item
curl -X POST http://localhost:8080/scheduled-items -H "Content-Type: application/json" -d "{\"title\":\"Test Scheduled Item\",\"description\":\"Test Description\",\"startsAt\":\"2023-05-20T15:00:00Z\",\"repeats\":false}"

echo.
echo 3. Get all scheduled items again (should include the new item)
curl -X GET http://localhost:8080/scheduled-items

echo.
echo 4. Get scheduled item with ID 3 (the one we just created)
curl -X GET http://localhost:8080/scheduled-items/3

echo.
echo 5. Update scheduled item with ID 3
curl -X PUT http://localhost:8080/scheduled-items/3 -H "Content-Type: application/json" -d "{\"title\":\"Updated Scheduled Item\",\"description\":\"Updated Description\",\"startsAt\":\"2023-05-21T16:30:00Z\",\"repeats\":true,\"cronExpression\":\"0 0 12 * * *\",\"expiration\":\"2023-12-31T23:59:59Z\"}"

echo.
echo 6. Get scheduled item with ID 3 again (should show updated values)
curl -X GET http://localhost:8080/scheduled-items/3

echo.
echo 7. Delete scheduled item with ID 3
curl -X DELETE http://localhost:8080/scheduled-items/3

echo.
echo 8. Get all scheduled items (should not include the deleted item)
curl -X GET http://localhost:8080/scheduled-items

echo.
echo Testing completed!
