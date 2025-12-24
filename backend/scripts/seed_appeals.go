package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"citizen-appeals/pkg/classification"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AppealTemplate struct {
	title       string
	description string
	address     string
	lat         float64
	lng         float64
	category    string // Назва категорії для пошуку
}

var (
	appealTemplates = []AppealTemplate{
		// ========== ВОДОКАНАЛ ==========
		{"Немає води в квартирі", "В квартирі немає води вже другий день. Неможливо нормально жити без водопостачання.", "вул. Лесі Українки, 25", 50.4520, 30.5250, "Інфраструктура та Мережі"},
		{"Прорвало водопровідну трубу", "На вул. Саксаганського прорвало водопровідну трубу. Вода заливає проїжджу частину та підвали будинків.", "вул. Саксаганського, 5", 50.4540, 30.5270, "Інфраструктура та Мережі"},
		{"Тече каналізація у дворі", "На вул. Бандери тече каналізація. Неприємний запах, небезпечно для здоров'я мешканців.", "вул. Бандери, 7", 50.4600, 30.5330, "Інфраструктура та Мережі"},
		{"Іржава вода з крана", "З крана тече іржава вода вже третій день. Неможливо пити та готувати їжу.", "вул. Хрещатик, 15", 50.4570, 30.5300, "Інфраструктура та Мережі"},
		{"Низький тиск води", "Дуже низький тиск води у будинку. Вода не доходить до верхніх поверхів.", "вул. Грушевського, 12", 50.4560, 30.5290, "Інфраструктура та Мережі"},
		{"Засмічена каналізація", "Каналізаційний люк засмічений, вода не відходить. Затоплює двір.", "вул. Львівська, 30", 50.4550, 30.5280, "Інфраструктура та Мережі"},

		// ========== ТЕПЛОПОСТАЧАННЯ ==========
		{"Холодні батареї в квартирі", "Батареї холодні у квартирі під час опалювального сезону. Мешканці мерзнуть, особливо діти.", "вул. Хрещатик, 12", 50.4570, 30.5300, "Інфраструктура та Мережі"},
		{"Прорив теплотраси", "Прорвало теплотрасу біля будинку. Гаряча вода та пара виходять на вулицю.", "вул. Шевченка, 20", 50.4510, 30.5240, "Інфраструктура та Мережі"},
		{"Немає опалення в будинку", "В будинку №15 не працює опалення вже тиждень. Температура в квартирах нижче норми.", "вул. Бандери, 15", 50.4530, 30.5260, "Інфраструктура та Мережі"},
		{"Низька температура в квартирі", "У квартирі дуже холодно, батареї ледь теплі. Діти постійно хворіють.", "вул. Саксаганського, 8", 50.4540, 30.5270, "Інфраструктура та Мережі"},
		{"Коли почнеться опалювальний сезон", "Вже листопад а опалення досі немає. Коли нарешті увімкнуть тепло?", "вул. Грушевського, 5", 50.4560, 30.5290, "Інфраструктура та Мережі"},

		// ========== ЕЛЕКТРОМЕРЕЖІ ==========
		{"Не горять ліхтарі на вулиці", "На вулиці Шевченка не горять ліхтарі вже тиждень. Дуже темно ввечері, небезпечно.", "вул. Шевченка, 20", 50.4510, 30.5240, "Інфраструктура та Мережі"},
		{"Немає світла в підʼїзді", "Не працює освітлення в підʼїзді будинку. Темно та небезпечно.", "вул. Бандери, 10", 50.4530, 30.5260, "Інфраструктура та Мережі"},
		{"Обірвався електричний провід", "На вулиці Грушевського обірвався електричний провід. Небезпечно для прохожих.", "вул. Грушевського, 3", 50.4560, 30.5290, "Інфраструктура та Мережі"},
		{"Постійні відключення електроенергії", "Постійно вимикають світло без попередження. Псується техніка.", "вул. Львівська, 25", 50.4550, 30.5280, "Інфраструктура та Мережі"},
		{"Пошкоджений електричний стовп", "Електричний стовп нахилився після бурі. Може впасти на дорогу.", "вул. Хрещатик, 8", 50.4570, 30.5300, "Інфраструктура та Мережі"},

		// ========== ГАЗОПОСТАЧАННЯ ==========
		{"Запах газу в підʼїзді", "Відчувається сильний запах газу в підʼїзді на першому поверсі. Терміново потрібна перевірка.", "вул. Грушевського, 8", 50.4560, 30.5290, "Інфраструктура та Мережі"},
		{"Не працює газова плита", "Газ не надходить до квартири. Плита не працює, неможливо готувати їжу.", "вул. Саксаганського, 15", 50.4540, 30.5270, "Інфраструктура та Мережі"},
		{"Витік газу з труби", "Підозра на витік газу з газової труби у дворі. Потрібна термінова перевірка.", "вул. Бандери, 4", 50.4530, 30.5260, "Інфраструктура та Мережі"},
		{"Проблеми з газовим котлом", "Газовий котел не працює. Немає гарячої води та опалення.", "вул. Лесі Українки, 19", 50.4520, 30.5250, "Інфраструктура та Мережі"},

		// ========== ОЗЕЛЕНЕННЯ ==========
		{"Аварійне дерево у дворі", "Біля будинку росте аварійне дерево. Може впасти на машини або людей.", "вул. Грушевського, 3", 50.4560, 30.5290, "Благоустрій та Довкілля"},
		{"Потрібна обрізка дерев", "Гілки дерева закривають вікна та чіпляють проводи. Потрібна обрізка.", "вул. Шевченка, 14", 50.4510, 30.5240, "Благоустрій та Довкілля"},
		{"Впало дерево після бурі", "Після вчорашньої бурі впало дерево у дворі. Перекриває прохід.", "вул. Хрещатик, 22", 50.4570, 30.5300, "Благоустрій та Довкілля"},
		{"Трава не кошена у сквері", "У сквері трава по пояс, ніхто не косить вже місяць. Виглядає занедбано.", "вул. Лесі Українки, 30", 50.4520, 30.5250, "Благоустрій та Довкілля"},
		{"Сухі гілки нависають над тротуаром", "Сухі гілки дерева нависають над тротуаром. Небезпечно для прохожих.", "вул. Бандери, 7", 50.4530, 30.5260, "Благоустрій та Довкілля"},

		// ========== ПРИБИРАННЯ ВУЛИЦЬ ==========
		{"Вулицю не прибирають після снігу", "Після снігопаду вулицю не прибирають. Дуже слизько, люди падають.", "вул. Саксаганського, 11", 50.4540, 30.5270, "Благоустрій та Довкілля"},
		{"Ожеледиця на тротуарі", "На тротуарі ожеледиця, ніхто не посипає. Вже кілька людей впали.", "вул. Львівська, 28", 50.4550, 30.5280, "Благоустрій та Довкілля"},
		{"Купи листя не прибрані", "У дворі купи листя лежать вже місяць. Ніхто не прибирає.", "вул. Грушевського, 8", 50.4560, 30.5290, "Благоустрій та Довкілля"},
		{"Двір брудний після ремонту", "Після ремонту дороги залишили бруд на тротуарі. Потрібне прибирання.", "вул. Хрещатик, 8", 50.4570, 30.5300, "Благоустрій та Довкілля"},
		{"Калюжі стоять на тротуарі", "Після дощу калюжі стоять на тротуарі тижнями. Вода не відводиться.", "вул. Шевченка, 25", 50.4510, 30.5240, "Благоустрій та Довкілля"},

		// ========== ВИВЕЗЕННЯ ТПВ ==========
		{"Сміття не вивозять", "Контейнери для сміття не вивозять вже 2 тижні. Сміття накопичилось, жахливий запах.", "вул. Львівська, 30", 50.4550, 30.5280, "Благоустрій та Довкілля"},
		{"Переповнені сміттєві баки", "Сміттєві баки переповнені, сміття розкидане по двору. Антисанітарія.", "вул. Бандери, 12", 50.4530, 30.5260, "Благоустрій та Довкілля"},
		{"Зламаний контейнер для сміття", "Контейнер для сміття зламаний, кришка не закривається. Собаки розтягують сміття.", "вул. Саксаганського, 20", 50.4540, 30.5270, "Благоустрій та Довкілля"},
		{"Порушений графік вивозу сміття", "Сміттєвоз не приїжджає за графіком. Постійно затримки на кілька днів.", "вул. Хрещатик, 18", 50.4570, 30.5300, "Благоустрій та Довкілля"},
		{"Потрібен додатковий контейнер", "У дворі один контейнер на весь будинок. Потрібен додатковий, не вистачає місця.", "вул. Грушевського, 15", 50.4560, 30.5290, "Благоустрій та Довкілля"},

		// ========== ГРОМАДСЬКИЙ ТРАНСПОРТ ==========
		{"Автобус не їде за розкладом", "Автобус 25 маршруту не їздить за розкладом. Чекаємо по 40 хвилин.", "вул. Лесі Українки, 19", 50.4520, 30.5250, "Транспорт"},
		{"Тролейбус довго не приїжджає", "Тролейбус №7 не зʼявляється вже годину. Люди мерзнуть на зупинці.", "вул. Шевченка, 30", 50.4510, 30.5240, "Транспорт"},
		{"Водій маршрутки хамить", "Водій маршрутки №523 хамить пасажирам та їде небезпечно.", "вул. Хрещатик, 1", 50.4570, 30.5300, "Транспорт"},
		{"Розбите скло на зупинці", "На зупинці громадського транспорту розбите скло. Небезпечно для людей.", "вул. Бандери, 10", 50.4530, 30.5260, "Транспорт"},
		{"Скасували маршрут без попередження", "Маршрут №45 скасували без попередження. Люди не можуть дістатися на роботу.", "вул. Грушевського, 22", 50.4560, 30.5290, "Транспорт"},

		// ========== МУНІЦИПАЛЬНА ВАРТА ==========
		{"Гучна музика після 22 години", "Сусіди слухають гучну музику після 22 години щодня. Неможливо спати.", "вул. Шевченка, 20", 50.4510, 30.5240, "Безпека та Порядок"},
		{"Шум від ремонту вночі", "Сусід робить ремонт вночі, шум до 3 години. Порушення тиші.", "вул. Бандери, 8", 50.4530, 30.5260, "Безпека та Порядок"},
		{"Пʼяна компанія шумить у дворі", "У дворі постійно збирається пʼяна компанія, шумлять та розбивають пляшки.", "вул. Саксаганського, 15", 50.4540, 30.5270, "Безпека та Порядок"},
		{"Хуліганство біля підʼїзду", "Група молоді постійно шумить біля підʼїзду, малюють графіті.", "вул. Львівська, 12", 50.4550, 30.5280, "Безпека та Порядок"},
		{"Бійка під будинком", "Під будинком постійно бʼються та сваряться. Страшно виходити ввечері.", "вул. Хрещатик, 25", 50.4570, 30.5300, "Безпека та Порядок"},
	}

	statuses = []string{"new", "assigned", "in_progress", "completed", "closed", "rejected"}
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Використання: go run seed_appeals.go <кількість_звернень>")
		fmt.Println("Приклад: go run seed_appeals.go 50")
		os.Exit(1)
	}

	var count int
	if _, err := fmt.Sscanf(os.Args[1], "%d", &count); err != nil || count <= 0 {
		log.Fatalf("Невірна кількість звернень: %s", os.Args[1])
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/citizen_appeals?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())

	// Ініціалізуємо класифікатор
	classificationURL := os.Getenv("CLASSIFICATION_SERVICE_URL")
	if classificationURL == "" {
		classificationURL = "http://localhost:8000"
	}
	classificationEnabled := os.Getenv("CLASSIFICATION_ENABLED")
	if classificationEnabled == "" {
		classificationEnabled = "true"
	}
	enabled := classificationEnabled == "true"
	classifier := classification.NewClassifier(classificationURL, enabled)

	if enabled {
		log.Printf("Класифікатор ініціалізовано: URL=%s, Enabled=%v", classificationURL, enabled)
	} else {
		log.Println("Класифікатор вимкнено, служби будуть призначатися випадково")
	}

	// Перевірка наявності даних
	var userCount, serviceCount, categoryCount int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount)
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM services WHERE is_active = true").Scan(&serviceCount)
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM categories WHERE is_active = true").Scan(&categoryCount)

	if userCount == 0 {
		log.Fatal("Немає користувачів в системі! Спочатку створи користувачів.")
	}
	if serviceCount == 0 {
		log.Fatal("Немає служб в системі! Спочатку створи служби.")
	}
	if categoryCount == 0 {
		log.Fatal("Немає категорій в системі!")
	}

	// Отримуємо ID користувачів різних ролей для історії
	var citizenID, dispatcherID, executorID int64
	err = pool.QueryRow(ctx, "SELECT id FROM users WHERE role = 'citizen' LIMIT 1").Scan(&citizenID)
	if err != nil {
		pool.QueryRow(ctx, "SELECT id FROM users LIMIT 1").Scan(&citizenID)
	}
	pool.QueryRow(ctx, "SELECT id FROM users WHERE role = 'dispatcher' LIMIT 1").Scan(&dispatcherID)
	pool.QueryRow(ctx, "SELECT id FROM users WHERE role = 'executor' LIMIT 1").Scan(&executorID)

	// Якщо немає dispatcher або executor, використовуємо будь-якого користувача
	if dispatcherID == 0 {
		dispatcherID = citizenID
	}
	if executorID == 0 {
		executorID = citizenID
	}

	// Отримуємо список категорій з назвами для маппінгу
	categoryRows, err := pool.Query(ctx, "SELECT id, name FROM categories WHERE is_active = true")
	if err != nil {
		log.Fatalf("Failed to get categories: %v", err)
	}
	defer categoryRows.Close()

	categoryMap := make(map[string]int64) // Назва категорії -> ID
	var categoryIDs []int64
	for categoryRows.Next() {
		var id int64
		var name string
		categoryRows.Scan(&id, &name)
		categoryMap[name] = id
		categoryIDs = append(categoryIDs, id)
	}

	// Отримуємо список служб
	serviceRows, err := pool.Query(ctx, "SELECT id FROM services WHERE is_active = true")
	if err != nil {
		log.Fatalf("Failed to get services: %v", err)
	}
	defer serviceRows.Close()

	var serviceIDs []int64
	for serviceRows.Next() {
		var id int64
		serviceRows.Scan(&id)
		serviceIDs = append(serviceIDs, id)
	}

	fmt.Printf("Створення %d тестових звернень...\n", count)

	// Генерація звернень
	for i := 0; i < count; i++ {
		// Вибираємо шаблон звернення (заголовок, опис, адреса та категорія пов'язані)
		template := appealTemplates[rand.Intn(len(appealTemplates))]

		// Знаходимо ID категорії за назвою
		categoryID, exists := categoryMap[template.category]
		if !exists {
			// Якщо категорія не знайдена, використовуємо випадкову
			if len(categoryIDs) == 0 {
				log.Printf("Помилка: немає категорій для звернення '%s'", template.title)
				continue
			}
			categoryID = categoryIDs[rand.Intn(len(categoryIDs))]
		}

		// Призначаємо службу ТІЛЬКИ через класифікацію
		var serviceID *int64
		// Використовуємо тільки опис для класифікації (заголовок часто занадто короткий)
		serviceName, confidence, err := classifier.ClassifyAppeal(ctx, template.description)
		if err != nil {
			if i < 5 {
				log.Printf("Помилка класифікації для звернення '%s': %v", template.title, err)
			}
		} else if serviceName != "" {
			// Шукаємо службу за назвою
			var foundServiceID int64
			err := pool.QueryRow(ctx, "SELECT id FROM services WHERE name = $1 AND is_active = true LIMIT 1", serviceName).Scan(&foundServiceID)
			if err == nil {
				serviceID = &foundServiceID
				if i < 5 || (i+1)%10 == 0 { // Логуємо перші 5 та кожне 10-те для невеликого шуму
					log.Printf("✓ Служба '%s' призначена через класифікацію (confidence: %.2f) для звернення '%s'", serviceName, confidence, template.title)
				}
			} else {
				if i < 5 {
					log.Printf("⚠ Служба '%s' з класифікації не знайдена в БД для звернення '%s'", serviceName, template.title)
				}
			}
		}
		// Якщо класифікація не повернула службу, залишаємо serviceID = nil (не призначаємо службу)

		// Розподіл статусів
		// Додаємо можливість створення звернень, закритих з простроченням
		status := statuses[rand.Intn(len(statuses))]
		randVal := rand.Float64()
		if randVal < 0.10 {
			status = "new"
		} else if randVal < 0.20 {
			status = "assigned"
		} else if randVal < 0.35 {
			status = "in_progress"
		} else if randVal < 0.55 {
			status = "completed"
		} else if randVal < 0.70 {
			status = "closed"
		} else if randVal < 0.85 {
			// Звернення, закриті з простроченням (completed після 30+ днів)
			status = "completed_overdue"
		} else if randVal < 0.95 {
			// Звернення, закриті з простроченням (closed після 30+ днів)
			status = "closed_overdue"
		} else {
			status = "rejected"
		}

		// Пріоритет
		priority := 2
		if rand.Float64() < 0.1 {
			priority = 3
		} else if rand.Float64() < 0.2 {
			priority = 1
		}

		// Дати (останні 90 днів)
		createdAt := time.Now().Add(-time.Duration(rand.Intn(90)) * 24 * time.Hour)

		// Генеруємо випадкові координати по центру Києва
		// Межі (скорочені вдвічі): широта 50.34-50.51, довгота 30.36-30.69
		kyivLatMin := 50.34
		kyivLatMax := 50.51
		kyivLngMin := 30.36
		kyivLngMax := 30.69

		randomLat := kyivLatMin + rand.Float64()*(kyivLatMax-kyivLatMin)
		randomLng := kyivLngMin + rand.Float64()*(kyivLngMax-kyivLngMin)

		// Зберігаємо оригінальний статус для generateStatusHistory
		originalStatus := status

		// Якщо служба призначена і статус 'new', змінюємо на 'assigned'
		if serviceID != nil && status == "new" {
			status = "assigned"
			originalStatus = "assigned"
		}

		// Конвертуємо спеціальні статуси для прострочених звернень перед вставкою в БД
		if status == "completed_overdue" {
			status = "completed"
		} else if status == "closed_overdue" {
			status = "closed"
		}

		query := `
			INSERT INTO appeals (user_id, category_id, service_id, status, title, description, address, latitude, longitude, priority, created_at, updated_at, closed_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			RETURNING id
		`

		// Спочатку вставляємо звернення з тимчасовими датами
		// Потім історія згенерує правильні дати, і ми оновимо updated_at та closed_at
		var appealID int64
		err = pool.QueryRow(ctx, query,
			citizenID, categoryID, serviceID, status, template.title, template.description,
			template.address, randomLat, randomLng, priority,
			createdAt, createdAt, nil, // Тимчасово встановлюємо updated_at = created_at
		).Scan(&appealID)

		if err != nil {
			log.Printf("Помилка створення звернення %d: %v", i+1, err)
			continue
		}

		// Генеруємо реалістичну історію статусів та отримуємо правильні дати
		// Використовуємо originalStatus для правильної обробки прострочених звернень
		lastHistoryDate, finalClosedAt := generateStatusHistory(ctx, pool, appealID, originalStatus, createdAt,
			citizenID, dispatcherID, executorID, serviceID)

		// Оновлюємо updated_at та closed_at на основі історії
		updateQuery := `UPDATE appeals SET updated_at = $1`
		updateArgs := []interface{}{lastHistoryDate}
		if finalClosedAt != nil {
			updateQuery += `, closed_at = $2`
			updateArgs = append(updateArgs, *finalClosedAt)
		}
		updateQuery += ` WHERE id = $` + fmt.Sprintf("%d", len(updateArgs)+1)
		updateArgs = append(updateArgs, appealID)

		pool.Exec(ctx, updateQuery, updateArgs...)
	}

	// Оновлюємо статус на 'assigned' для звернень з призначеною службою та статусом 'new'
	// (це має бути вже зроблено при створенні, але на всяк випадок)
	pool.Exec(ctx, `
		UPDATE appeals
		SET status = 'assigned'
		WHERE service_id IS NOT NULL
			AND status = 'new'
	`)

	var total int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM appeals").Scan(&total)
	fmt.Printf("✅ Готово! Створено звернень. Всього в БД: %d\n", total)
}

// generateStatusHistory створює реалістичну історію переходів статусів
// Повертає дату останньої зміни статусу та closed_at (якщо потрібно)
func generateStatusHistory(ctx context.Context, pool *pgxpool.Pool, appealID int64, finalStatus string,
	createdAt time.Time,
	citizenID, dispatcherID, executorID int64, serviceID *int64) (time.Time, *time.Time) {

	historyQuery := `
		INSERT INTO appeal_history (appeal_id, user_id, old_status, new_status, action, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	statusLabels := map[string]string{
		"new":         "Нове",
		"assigned":    "Призначене",
		"in_progress": "В роботі",
		"completed":   "Виконане",
		"closed":      "Закрите",
		"rejected":    "Відхилене",
	}

	currentDate := createdAt
	historyEntries := []struct {
		oldStatus string
		newStatus string
		userID    int64
		date      time.Time
		action    string
	}{}

	// Визначаємо послідовність переходів на основі фінального статусу
	var lastHistoryDate time.Time = createdAt
	var finalClosedAt *time.Time

	switch finalStatus {
	case "new":
		// Залишається новим, історії немає
		return createdAt, nil

	case "completed_overdue":
		// Звернення, закриті з простроченням (completed після 30+ днів)
		if serviceID != nil {
			// new -> assigned -> in_progress -> completed (після 30+ днів)
			assignedDate := currentDate.Add(time.Duration(rand.Intn(2)+1) * time.Hour)
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"new", "assigned", dispatcherID, assignedDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["new"], statusLabels["assigned"]),
			})

			inProgressDate := assignedDate.Add(time.Duration(rand.Intn(24)+1) * time.Hour)
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"assigned", "in_progress", executorID, inProgressDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["assigned"], statusLabels["in_progress"]),
			})

			// completed через 31-60 днів після початку роботи (прострочення)
			completedDate := inProgressDate.Add(time.Duration(rand.Intn(30)+31) * 24 * time.Hour)
			// Переконаємося, що дата не в майбутньому
			if completedDate.After(time.Now()) {
				completedDate = time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour)
			}
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"in_progress", "completed", executorID, completedDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["in_progress"], statusLabels["completed"]),
			})
			lastHistoryDate = completedDate
			finalClosedAt = &completedDate
		} else {
			// new -> completed (після 30+ днів, рідкісний випадок)
			completedDate := currentDate.Add(time.Duration(rand.Intn(30)+31) * 24 * time.Hour)
			// Переконаємося, що дата не в майбутньому
			if completedDate.After(time.Now()) {
				completedDate = time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour)
			}
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"new", "completed", dispatcherID, completedDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["new"], statusLabels["completed"]),
			})
			lastHistoryDate = completedDate
			finalClosedAt = &completedDate
		}

	case "closed_overdue":
		// Звернення, закриті з простроченням (closed після 30+ днів)
		if serviceID != nil {
			// new -> assigned -> in_progress -> completed -> closed (після 30+ днів)
			assignedDate := currentDate.Add(time.Duration(rand.Intn(2)+1) * time.Hour)
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"new", "assigned", dispatcherID, assignedDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["new"], statusLabels["assigned"]),
			})

			inProgressDate := assignedDate.Add(time.Duration(rand.Intn(24)+1) * time.Hour)
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"assigned", "in_progress", executorID, inProgressDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["assigned"], statusLabels["in_progress"]),
			})

			// completed через 31-50 днів після початку роботи (прострочення)
			completedDate := inProgressDate.Add(time.Duration(rand.Intn(20)+31) * 24 * time.Hour)
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"in_progress", "completed", executorID, completedDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["in_progress"], statusLabels["completed"]),
			})

			// closed через 1-3 дні після completed
			closedDate := completedDate.Add(time.Duration(rand.Intn(3)+1) * 24 * time.Hour)
			// Переконаємося, що дата не в майбутньому
			if closedDate.After(time.Now()) {
				closedDate = time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour)
			}
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"completed", "closed", dispatcherID, closedDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["completed"], statusLabels["closed"]),
			})
			lastHistoryDate = closedDate
			finalClosedAt = &closedDate
		} else {
			// new -> closed (після 30+ днів, рідкісний випадок)
			closedDate := currentDate.Add(time.Duration(rand.Intn(20)+31) * 24 * time.Hour)
			// Переконаємося, що дата не в майбутньому
			if closedDate.After(time.Now()) {
				closedDate = time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour)
			}
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"new", "closed", dispatcherID, closedDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["new"], statusLabels["closed"]),
			})
			lastHistoryDate = closedDate
			finalClosedAt = &closedDate
		}

	case "assigned":
		if serviceID != nil {
			// new -> assigned (диспетчер призначає)
			historyDate := currentDate.Add(time.Duration(rand.Intn(2)+1) * time.Hour)
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"new", "assigned", dispatcherID, historyDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["new"], statusLabels["assigned"]),
			})
			lastHistoryDate = historyDate
		}

	case "in_progress":
		if serviceID != nil {
			// new -> assigned -> in_progress
			assignedDate := currentDate.Add(time.Duration(rand.Intn(2)+1) * time.Hour)
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"new", "assigned", dispatcherID, assignedDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["new"], statusLabels["assigned"]),
			})

			inProgressDate := assignedDate.Add(time.Duration(rand.Intn(24)+1) * time.Hour)
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"assigned", "in_progress", executorID, inProgressDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["assigned"], statusLabels["in_progress"]),
			})
			lastHistoryDate = inProgressDate
		} else {
			// new -> in_progress (без призначення, рідкісний випадок)
			inProgressDate := currentDate.Add(time.Duration(rand.Intn(24)+1) * time.Hour)
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"new", "in_progress", dispatcherID, inProgressDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["new"], statusLabels["in_progress"]),
			})
			lastHistoryDate = inProgressDate
		}

	case "completed":
		if serviceID != nil {
			// new -> assigned -> in_progress -> completed
			assignedDate := currentDate.Add(time.Duration(rand.Intn(2)+1) * time.Hour)
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"new", "assigned", dispatcherID, assignedDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["new"], statusLabels["assigned"]),
			})

			inProgressDate := assignedDate.Add(time.Duration(rand.Intn(24)+1) * time.Hour)
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"assigned", "in_progress", executorID, inProgressDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["assigned"], statusLabels["in_progress"]),
			})

			// completed може бути через 1-20 днів після початку роботи
			completedDate := inProgressDate.Add(time.Duration(rand.Intn(20)+1) * 24 * time.Hour)
			// Переконаємося, що дата не в майбутньому
			if completedDate.After(time.Now()) {
				completedDate = time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour)
			}
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"in_progress", "completed", executorID, completedDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["in_progress"], statusLabels["completed"]),
			})
			lastHistoryDate = completedDate
			finalClosedAt = &completedDate
		} else {
			// new -> completed (рідкісний випадок, без призначення)
			completedDate := currentDate.Add(time.Duration(rand.Intn(5)+1) * 24 * time.Hour)
			// Переконаємося, що дата не в майбутньому
			if completedDate.After(time.Now()) {
				completedDate = time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour)
			}
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"new", "completed", dispatcherID, completedDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["new"], statusLabels["completed"]),
			})
			lastHistoryDate = completedDate
			finalClosedAt = &completedDate
		}

	case "closed":
		if serviceID != nil {
			// new -> assigned -> in_progress -> completed -> closed
			assignedDate := currentDate.Add(time.Duration(rand.Intn(2)+1) * time.Hour)
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"new", "assigned", dispatcherID, assignedDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["new"], statusLabels["assigned"]),
			})

			inProgressDate := assignedDate.Add(time.Duration(rand.Intn(24)+1) * time.Hour)
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"assigned", "in_progress", executorID, inProgressDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["assigned"], statusLabels["in_progress"]),
			})

			completedDate := inProgressDate.Add(time.Duration(rand.Intn(20)+1) * 24 * time.Hour)
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"in_progress", "completed", executorID, completedDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["in_progress"], statusLabels["completed"]),
			})

			// closed через 1-3 дні після completed
			closedDate := completedDate.Add(time.Duration(rand.Intn(3)+1) * 24 * time.Hour)
			// Переконаємося, що дата не в майбутньому
			if closedDate.After(time.Now()) {
				closedDate = time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour)
			}
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"completed", "closed", dispatcherID, closedDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["completed"], statusLabels["closed"]),
			})
			lastHistoryDate = closedDate
			finalClosedAt = &closedDate
		} else {
			// new -> closed (рідкісний випадок)
			closedDate := currentDate.Add(time.Duration(rand.Intn(5)+1) * 24 * time.Hour)
			// Переконаємося, що дата не в майбутньому
			if closedDate.After(time.Now()) {
				closedDate = time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour)
			}
			historyEntries = append(historyEntries, struct {
				oldStatus string
				newStatus string
				userID    int64
				date      time.Time
				action    string
			}{
				"new", "closed", dispatcherID, closedDate,
				fmt.Sprintf("Статус змінено з %s на %s", statusLabels["new"], statusLabels["closed"]),
			})
			lastHistoryDate = closedDate
			finalClosedAt = &closedDate
		}

	case "rejected":
		// Відхилення може статися на будь-якому етапі
		var rejectDate time.Time
		var rejectUserID int64
		var oldStatusForReject string

		if serviceID != nil {
			// Можливі сценарії відхилення
			scenario := rand.Float64()
			if scenario < 0.3 {
				// Відхилено одразу після призначення
				assignedDate := currentDate.Add(time.Duration(rand.Intn(2)+1) * time.Hour)
				historyEntries = append(historyEntries, struct {
					oldStatus string
					newStatus string
					userID    int64
					date      time.Time
					action    string
				}{
					"new", "assigned", dispatcherID, assignedDate,
					fmt.Sprintf("Статус змінено з %s на %s", statusLabels["new"], statusLabels["assigned"]),
				})
				rejectDate = assignedDate.Add(time.Duration(rand.Intn(24)+1) * time.Hour)
				oldStatusForReject = "assigned"
				rejectUserID = dispatcherID
				lastHistoryDate = rejectDate
			} else if scenario < 0.6 {
				// Відхилено після початку роботи
				assignedDate := currentDate.Add(time.Duration(rand.Intn(2)+1) * time.Hour)
				historyEntries = append(historyEntries, struct {
					oldStatus string
					newStatus string
					userID    int64
					date      time.Time
					action    string
				}{
					"new", "assigned", dispatcherID, assignedDate,
					fmt.Sprintf("Статус змінено з %s на %s", statusLabels["new"], statusLabels["assigned"]),
				})

				inProgressDate := assignedDate.Add(time.Duration(rand.Intn(24)+1) * time.Hour)
				historyEntries = append(historyEntries, struct {
					oldStatus string
					newStatus string
					userID    int64
					date      time.Time
					action    string
				}{
					"assigned", "in_progress", executorID, inProgressDate,
					fmt.Sprintf("Статус змінено з %s на %s", statusLabels["assigned"], statusLabels["in_progress"]),
				})
				rejectDate = inProgressDate.Add(time.Duration(rand.Intn(3)+1) * 24 * time.Hour)
				// Переконаємося, що дата не в майбутньому
				if rejectDate.After(time.Now()) {
					rejectDate = time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour)
				}
				oldStatusForReject = "in_progress"
				rejectUserID = executorID
				lastHistoryDate = rejectDate
			} else {
				// Відхилено одразу
				rejectDate = currentDate.Add(time.Duration(rand.Intn(24)+1) * time.Hour)
				// Переконаємося, що дата не в майбутньому
				if rejectDate.After(time.Now()) {
					rejectDate = time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour)
				}
				oldStatusForReject = "new"
				rejectUserID = dispatcherID
				lastHistoryDate = rejectDate
			}
		} else {
			// Відхилено без призначення
			rejectDate = currentDate.Add(time.Duration(rand.Intn(24)+1) * time.Hour)
			// Переконаємося, що дата не в майбутньому
			if rejectDate.After(time.Now()) {
				rejectDate = time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour)
			}
			oldStatusForReject = "new"
			rejectUserID = dispatcherID
			lastHistoryDate = rejectDate
		}

		historyEntries = append(historyEntries, struct {
			oldStatus string
			newStatus string
			userID    int64
			date      time.Time
			action    string
		}{
			oldStatusForReject, "rejected", rejectUserID, rejectDate,
			fmt.Sprintf("Статус змінено з %s на %s", statusLabels[oldStatusForReject], statusLabels["rejected"]),
		})
	}

	// Переконаємося, що lastHistoryDate не менше за createdAt
	if lastHistoryDate.Before(createdAt) {
		lastHistoryDate = createdAt
	}

	// Вставляємо всі записи історії в хронологічному порядку
	// Сортуємо за датою, щоб переконатися, що вони в правильному порядку
	for i := 0; i < len(historyEntries); i++ {
		for j := i + 1; j < len(historyEntries); j++ {
			if historyEntries[i].date.After(historyEntries[j].date) {
				historyEntries[i], historyEntries[j] = historyEntries[j], historyEntries[i]
			}
		}
	}

	// Перевіряємо хронологічність та виправляємо, якщо потрібно
	now := time.Now()
	for i := 0; i < len(historyEntries); i++ {
		// Визначаємо мінімальну дату для цього запису
		minDate := createdAt
		if i > 0 {
			// Мінімальна дата - це попередня дата + мінімум 1 година
			minDate = historyEntries[i-1].date.Add(1 * time.Hour)
		}

		// Визначаємо максимальну дату для цього запису
		maxDate := now
		if i < len(historyEntries)-1 {
			// Максимальна дата - це наступна дата - мінімум 1 година
			maxDate = historyEntries[i+1].date.Add(-1 * time.Hour)
		}

		// Виправляємо дату, якщо вона виходить за межі
		if historyEntries[i].date.Before(minDate) {
			// Якщо дата раніше за мінімальну, встановлюємо мінімальну + випадковий інтервал
			if maxDate.After(minDate) {
				diff := maxDate.Sub(minDate)
				if diff > 24*time.Hour {
					diff = 24 * time.Hour
				}
				diffHours := int(diff.Hours())
				if diffHours > 0 {
					historyEntries[i].date = minDate.Add(time.Duration(rand.Intn(diffHours)) * time.Hour)
				} else {
					historyEntries[i].date = minDate
				}
			} else {
				historyEntries[i].date = minDate
			}
		}

		if historyEntries[i].date.After(maxDate) {
			// Якщо дата пізніше за максимальну, встановлюємо максимальну - випадковий інтервал
			if maxDate.After(minDate) {
				diff := maxDate.Sub(minDate)
				if diff > 24*time.Hour {
					diff = 24 * time.Hour
				}
				diffHours := int(diff.Hours())
				if diffHours > 0 {
					historyEntries[i].date = maxDate.Add(-time.Duration(rand.Intn(diffHours)) * time.Hour)
				} else {
					historyEntries[i].date = maxDate
				}
			} else {
				historyEntries[i].date = maxDate
			}
		}

		// Переконаємося, що дата не в майбутньому
		if historyEntries[i].date.After(now) {
			historyEntries[i].date = now.Add(-time.Duration(rand.Intn(24)) * time.Hour)
		}

		// Переконаємося, що між записами є мінімальний інтервал (1 година)
		if i > 0 {
			diff := historyEntries[i].date.Sub(historyEntries[i-1].date)
			if diff < 1*time.Hour {
				historyEntries[i].date = historyEntries[i-1].date.Add(1 * time.Hour)
			}
		}
	}

	// Оновлюємо lastHistoryDate на основі останнього запису
	if len(historyEntries) > 0 {
		lastEntry := historyEntries[len(historyEntries)-1]
		lastHistoryDate = lastEntry.date
	}

	// Вставляємо всі записи історії
	for _, entry := range historyEntries {
		_, err := pool.Exec(ctx, historyQuery,
			appealID, entry.userID, entry.oldStatus, entry.newStatus, entry.action, entry.date)
		if err != nil {
			log.Printf("Помилка додавання історії для звернення %d: %v", appealID, err)
		}
	}

	return lastHistoryDate, finalClosedAt
}

func getStatusLabel(status string) string {
	labels := map[string]string{
		"new":         "Нове",
		"assigned":    "Призначене",
		"in_progress": "В роботі",
		"completed":   "Виконане",
		"closed":      "Закрите",
		"rejected":    "Відхилене",
	}
	return labels[status]
}
