# 🔒 Тесты безопасности Secure File Manager

Эта папка содержит автоматические тесты для проверки защиты от уязвимостей.

## 📋 Список тестов

| Файл | Уязвимость | Что проверяет |
|------|------------|---------------|
| `path_traversal_test.go` | Path Traversal | Попытки `../`, абсолютные пути |
| `zip_attacks_test.go` | ZIP Bomb, Zip Slip | Архивы-бомбы, path traversal в ZIP |
| `race_condition_test.go` | Race Condition | Параллельный доступ к файлам |
| `sql_injection_test.go` | SQL Injection | Prepared Statements, плейсхолдеры |
| `deserialization_test.go` | Insecure Deserialization | JSON/XML парсинг, XXE |

---

## 🚀 Как запустить тесты

### Все тесты сразу
```bash
go test -v ./tests/...
```

### Отдельные тесты по категориям

```bash
# Path Traversal
go test -v ./tests/... -run TestPathTraversal

# ZIP атаки (бомбы и Zip Slip)
go test -v ./tests/... -run TestZip

# Race Condition
go test -v ./tests/... -run TestRaceCondition

# SQL Injection
go test -v ./tests/... -run TestSQLInjection

# Десериализация
go test -v ./tests/... -run TestInsecureDeserialization
```

### Сохранить результаты в файл
```bash
go test -v ./tests/... > security_report.txt 2>&1
```

---

## 📊 Пример вывода

```
=== RUN   TestPathTraversal
=== RUN   TestPathTraversal/Attack_ParentDir
    path_traversal_test.go:78: ✅ ЗАЩИТА РАБОТАЕТ: Попытка выйти на уровень выше
=== RUN   TestPathTraversal/Attack_DeepTraversal
    path_traversal_test.go:78: ✅ ЗАЩИТА РАБОТАЕТ: Попытка добраться до /etc/passwd
--- PASS: TestPathTraversal

=== RUN   TestZipSlipProtection
=== RUN   TestZipSlipProtection/DeepTraversal
    zip_attacks_test.go:82: ✅ ЗАЩИТА ОТ ZIP SLIP: Попытка записи в системную папку
--- PASS: TestZipSlipProtection

=== RUN   TestSQLInjectionProtection/PreparedStatements
    sql_injection_test.go:36: ✅ users.go: Используются Prepared Statements
    sql_injection_test.go:40: ✅ users.go: Используются плейсхолдеры PostgreSQL
--- PASS: TestSQLInjectionProtection
```

---

## ✅ Ожидаемые результаты

| Тест | Ожидаемый результат |
|------|---------------------|
| Path Traversal | Все атаки заблокированы |
| Zip Slip | Все вредоносные пути заблокированы |
| ZIP Bomb | Показывает наличие защиты в коде |
| Race Condition | Файлы не повреждаются при параллельном доступе |
| SQL Injection | Все файлы используют Prepared Statements |
| Deserialization | Код не выполняется при парсинге JSON/XML |

---

## 🔧 Добавление новых тестов

Создайте файл `tests/vulnerability_name_test.go`:

```go
package tests

import "testing"

func TestNewVulnerability(t *testing.T) {
    t.Run("AttackScenario", func(t *testing.T) {
        // Ваш код теста
        if vulnerabilityExists {
            t.Error("❌ УЯЗВИМОСТЬ!")
        } else {
            t.Log("✅ Защита работает")
        }
    })
}
```
