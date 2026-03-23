## Done! Here's the Postman JSON for testing:

1. Register Push Token
   POST /api/notifications/push/register
   {
   "push_token": "ExponentPushToken[xxxxxxxxxxxxxxx]",
   "device": "android"
   }

---

2. Send Notification to User
   POST /api/notifications/notify
   {
   "user_id": 1,
   "title": "New Message",
   "body": "You have a new message from John",
   "type": "chat",
   "data": {
   "chat_id": 123,
   "sender_id": 2
   }
   }

---

3. Broadcast (All Users)
   POST /api/notifications/broadcast
   {
   "title": "System Maintenance",
   "body": "Scheduled maintenance at 2:00 AM",
   "type": "system",
   "data": {
   "maintenance_time": "2026-03-22T02:00:00Z"
   }
   }

---

4. Get My Notifications
   GET /api/notifications/

---

5. Get Unread Count
   GET /api/notifications/unread-count

---

6. Mark Single as Read
   PUT /api/notifications/1/read

---

7. Mark All as Read
   PUT /api/notifications/read-all

---

8. Unregister Push Token
   POST /api/notifications/push/unregister
   {
   "push_token": "ExponentPushToken[xxxxxxxxxxxxxxx]"
   }

---

Note: Requires Authorization: Bearer <token> header for authenticated endpoints.
