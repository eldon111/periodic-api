package integration

import (
	"periodic-api/internal/models"
	"periodic-api/internal/store"
	"bytes"
	"testing"
)

func TestUserIntegration(t *testing.T) {
	skipIfDBNotAvailable(t)

	// Clean up before and after the test
	cleanupUsers(t)
	defer cleanupUsers(t)

	// Create store
	userStore := store.NewPostgresUserStore(getActiveDB())

	// Test data
	testUser := models.User{
		Username:     "integration_test_user",
		PasswordHash: []byte("hashed_password_123"),
	}

	t.Run("Full CRUD Workflow", func(t *testing.T) {
		// Create
		created := userStore.CreateUser(testUser)
		if created.ID == 0 {
			t.Fatal("Created user should have non-zero ID")
		}
		if created.Username != testUser.Username {
			t.Errorf("Expected username %s, got %s", testUser.Username, created.Username)
		}
		if !bytes.Equal(created.PasswordHash, testUser.PasswordHash) {
			t.Error("Password hash should match")
		}

		// Read
		retrieved, found := userStore.GetUser(created.ID)
		if !found {
			t.Fatal("Should find the created user")
		}
		if retrieved.Username != testUser.Username {
			t.Errorf("Expected username %s, got %s", testUser.Username, retrieved.Username)
		}
		if !bytes.Equal(retrieved.PasswordHash, testUser.PasswordHash) {
			t.Error("Password hash should match after retrieval")
		}

		// Update
		updated := models.User{
			Username:     "updated_integration_user",
			PasswordHash: []byte("new_hashed_password_456"),
		}

		result, success := userStore.UpdateUser(created.ID, updated)
		if !success {
			t.Fatal("Update should succeed")
		}
		if result.Username != updated.Username {
			t.Errorf("Expected updated username %s, got %s", updated.Username, result.Username)
		}
		if !bytes.Equal(result.PasswordHash, updated.PasswordHash) {
			t.Error("Updated password hash should match")
		}

		// Verify update persisted
		verified, found := userStore.GetUser(created.ID)
		if !found {
			t.Fatal("Should still find the user after update")
		}
		if verified.Username != updated.Username {
			t.Errorf("Update should persist: expected username %s, got %s", updated.Username, verified.Username)
		}
		if !bytes.Equal(verified.PasswordHash, updated.PasswordHash) {
			t.Error("Updated password hash should persist")
		}

		// Delete
		deleted := userStore.DeleteUser(created.ID)
		if !deleted {
			t.Fatal("Delete should succeed")
		}

		// Verify deletion
		_, found = userStore.GetUser(created.ID)
		if found {
			t.Error("User should not be found after deletion")
		}
	})

	t.Run("Multiple Users Operations", func(t *testing.T) {
		// Create multiple users
		users := []models.User{
			{Username: "user1", PasswordHash: []byte("hash1")},
			{Username: "user2", PasswordHash: []byte("hash2")},
			{Username: "user3", PasswordHash: []byte("hash3")},
		}

		var createdIDs []int64
		for _, user := range users {
			created := userStore.CreateUser(user)
			createdIDs = append(createdIDs, created.ID)
		}

		// Get all users
		allUsers := userStore.GetAllUsers()
		if len(allUsers) < 3 {
			t.Errorf("Expected at least 3 users, got %d", len(allUsers))
		}

		// Verify our users are in the list
		foundCount := 0
		foundUsernames := make(map[string]bool)
		for _, user := range allUsers {
			for _, id := range createdIDs {
				if user.ID == id {
					foundCount++
					foundUsernames[user.Username] = true
					break
				}
			}
		}
		if foundCount != 3 {
			t.Errorf("Expected to find 3 created users, found %d", foundCount)
		}

		// Verify specific usernames
		expectedUsernames := []string{"user1", "user2", "user3"}
		for _, username := range expectedUsernames {
			if !foundUsernames[username] {
				t.Errorf("Expected to find username %s", username)
			}
		}

		// Clean up
		for _, id := range createdIDs {
			userStore.DeleteUser(id)
		}
	})

	t.Run("Username Uniqueness", func(t *testing.T) {
		// Create first user
		user1 := models.User{
			Username:     "unique_test_user",
			PasswordHash: []byte("password1"),
		}
		created1 := userStore.CreateUser(user1)

		// Try to create user with same username
		user2 := models.User{
			Username:     "unique_test_user",
			PasswordHash: []byte("password2"),
		}
		created2 := userStore.CreateUser(user2)

		// Both should be created (no unique constraint in current schema)
		// But they should have different IDs
		if created1.ID == created2.ID {
			t.Error("Different users should have different IDs")
		}

		// Clean up
		userStore.DeleteUser(created1.ID)
		userStore.DeleteUser(created2.ID)
	})

	t.Run("Password Hash Handling", func(t *testing.T) {
		// Test with different password hash sizes
		testCases := []struct {
			name string
			hash []byte
		}{
			{"Empty hash", []byte{}},
			{"Short hash", []byte("short")},
			{"Long hash", []byte("this_is_a_very_long_password_hash_that_might_be_used_in_real_applications")},
			{"Binary hash", []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD}},
		}

		var createdIDs []int64
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				user := models.User{
					Username:     "hash_test_" + tc.name,
					PasswordHash: tc.hash,
				}
				created := userStore.CreateUser(user)
				createdIDs = append(createdIDs, created.ID)

				if !bytes.Equal(created.PasswordHash, tc.hash) {
					t.Errorf("Password hash not preserved during creation")
				}

				// Verify retrieval
				retrieved, found := userStore.GetUser(created.ID)
				if !found {
					t.Error("Should find created user")
				}
				if !bytes.Equal(retrieved.PasswordHash, tc.hash) {
					t.Errorf("Password hash not preserved during retrieval")
				}
			})
		}

		// Clean up
		for _, id := range createdIDs {
			userStore.DeleteUser(id)
		}
	})

	t.Run("Edge Cases", func(t *testing.T) {
		// Test with non-existent ID
		_, found := userStore.GetUser(99999)
		if found {
			t.Error("Should not find non-existent user")
		}

		// Test update non-existent user
		_, success := userStore.UpdateUser(99999, testUser)
		if success {
			t.Error("Update of non-existent user should fail")
		}

		// Test delete non-existent user
		deleted := userStore.DeleteUser(99999)
		if deleted {
			t.Error("Delete of non-existent user should fail")
		}

		// Test with empty username (should still work)
		emptyUser := models.User{
			Username:     "",
			PasswordHash: []byte("some_hash"),
		}
		created := userStore.CreateUser(emptyUser)
		if created.ID == 0 {
			t.Error("Should be able to create user with empty username")
		}
		userStore.DeleteUser(created.ID)
	})
}

func cleanupUsers(t *testing.T) {
	_, err := getActiveDB().Exec("DELETE FROM users")
	if err != nil {
		t.Logf("Failed to cleanup users: %v", err)
	}
}
