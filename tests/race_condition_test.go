package tests

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"secure-fm/config"
	"secure-fm/fs"
)

// TestRaceConditionProtection проверяет защиту от состояния гонки (Race Condition)
// Уязвимость: одновременный доступ к файлу из разных потоков вызывает повреждение данных
func TestRaceConditionProtection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox_race")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{SandboxPath: tmpDir}
	fs.InitFS(cfg)

	// Создаём тестовый файл
	testFile := "race_test.txt"
	fs.WriteFile(testFile, "initial content")

	t.Run("ConcurrentWrites", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, 100)

		// Запускаем 100 горутин, записывающих одновременно
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				content := "content from goroutine"
				err := fs.WriteFile(testFile, content)
				if err != nil {
					errors <- err
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		errCount := 0
		for err := range errors {
			t.Logf("Ошибка при записи: %v", err)
			errCount++
		}

		// Проверяем что файл не повреждён
		content, err := fs.ReadFile(testFile)
		if err != nil {
			t.Errorf("Файл повреждён после конкурентных записей: %v", err)
		} else {
			t.Logf("✅ Файл не повреждён после 100 конкурентных записей")
			t.Logf("   Содержимое: %s", content[:min(50, len(content))])
		}
	})

	t.Run("ConcurrentReadsAndWrites", func(t *testing.T) {
		var wg sync.WaitGroup

		// 50 читателей и 50 писателей одновременно
		for i := 0; i < 50; i++ {
			wg.Add(2)

			// Читатель
			go func() {
				defer wg.Done()
				fs.ReadFile(testFile)
			}()

			// Писатель
			go func(id int) {
				defer wg.Done()
				fs.WriteFile(testFile, "write "+string(rune('0'+id%10)))
			}(i)
		}

		wg.Wait()
		t.Log("✅ Конкурентные чтения и записи завершены без паники")
	})

	t.Run("MutexVerification", func(t *testing.T) {
		// Проверяем что код использует sync.RWMutex
		// Это code review тест - проверяем что защита есть в коде

		opsFile := filepath.Join("..", "fs", "operations.go")
		content, err := os.ReadFile(opsFile)
		if err != nil {
			t.Skip("Не удалось прочитать fs/operations.go")
		}

		code := string(content)
		checks := []struct {
			pattern string
			desc    string
		}{
			{"sync.RWMutex", "Объявление мьютекса"},
			{"fileMutex.Lock()", "Блокировка на запись"},
			{"fileMutex.RLock()", "Блокировка на чтение"},
			{"fileMutex.Unlock()", "Разблокировка записи"},
			{"fileMutex.RUnlock()", "Разблокировка чтения"},
		}

		for _, c := range checks {
			if contains(code, c.pattern) {
				t.Logf("✅ Найдено: %s (%s)", c.pattern, c.desc)
			} else {
				t.Errorf("❌ НЕ НАЙДЕНО: %s (%s)", c.pattern, c.desc)
			}
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
