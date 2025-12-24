package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ServiceEmbedding struct {
	ServiceID   int64     `bson:"service_id"`
	ServiceName string    `bson:"service_name"`
	Description string    `bson:"description"`
	Examples    []string  `bson:"examples"`
	Version     int       `bson:"version"`
	UpdatedAt   time.Time `bson:"updated_at"`
	CreatedAt   time.Time `bson:"created_at"`
}

// Дані служб з міграцій
var servicesData = []struct {
	Name        string
	Description string
	Keywords    string
}{
	{
		Name:        "Водоканал",
		Description: "Вода, водопостачання, каналізація, прорив труб, якість води, аварії на водопроводі",
		Keywords:    "Немає води у квартирі вже другий день; Прорвало трубу у дворі та затоплює підвал; З каналізаційного люка тече вода; Низький тиск води у будинку; Вода з крана йде брудна жовтого кольору; Іржава вода тече з крана; Затоплення підвалу через прорив водопроводу; Каналізаційна яма переповнена та витікає; Вода не доходить до верхніх поверхів; Прорив труби на вулиці вода тече на дорогу; Запах каналізації у підʼїзді; Вода з крана має дивний смак; Каналізаційний колодязь засмічений; Водопровідна труба пошкоджена потрібен ремонт; Немає води в цілому будинку; Забита каналізація у дворі; Тече вода з люка на тротуар; Постійно тече вода у підвалі; Довго немає гарячої води",
	},
	{
		Name:        "Теплопостачання",
		Description: "Опалення, гаряча вода, теплові мережі, котельні, температура в квартирі",
		Keywords:    "Батареї холодні у квартирі під час опалювального сезону; Не працює опалення у будинку; Прорвало теплотрасу біля будинку; У квартирі дуже низька температура; Холодні батареї вже тиждень; Батареї ледь теплі а надворі мороз; Немає гарячої води; Опалювальний сезон почався а тепла немає; Прорив теплотраси вода хлеще на вулицю; У квартирі холодно діти хворіють; Не гріють батареї в кінці будинку; Температура в квартирі нижче норми; Коли почнеться опалювальний сезон; Тепла немає вже місяць; Котельня не працює будинок холодний",
	},
	{
		Name:        "Електромережі",
		Description: "Світло, електрика, вуличне освітлення, обриви проводів, трансформатори, напруга",
		Keywords:    "Немає світла у підʼїзді; Не працює вуличне освітлення; Обірвався електричний провід; Часті відключення електроенергії; Не горять ліхтарі на вулиці вже тиждень; Дуже темно ввечері на вулиці; Стрибки напруги псують техніку; Іскрить електрична проводка; Постійно вимикають світло; Немає освітлення біля школи; Не горить лампа над входом; Пошкоджений електричний стовп; Обрив лінії електропередач; Небезпечний провід висить низько; Вночі повна темрява ліхтарі не працюють",
	},
	{
		Name:        "Газопостачання",
		Description: "Газ, витік газу, запах газу, газова труба, аварійна газова служба",
		Keywords:    "Відчувається сильний запах газу у квартирі; Витік газу з газової труби; Аварійна ситуація з газопостачанням; Не працює газовий котел; Запах газу в підʼїзді на першому поверсі; Газова труба пошкоджена; Тхне газом біля будинку; Не працює газова плита; Газ не горить синім полумʼям; Відчувається запах газу на вулиці; Проблеми з газовим лічильником; Потрібна перевірка газового обладнання; Підозра на витік газу в підвалі; Не можу увімкнути газову колонку; Тисне з газової труби",
	},
	{
		Name:        "Озеленення",
		Description: "Дерева, парки, газони, обрізка гілок, аварійні дерева, сквери",
		Keywords:    "Потрібно спиляти аварійне дерево; Впало дерево у дворі; Сухі гілки нависають над дорогою; Не доглянута зелена зона; Дерево може впасти на машини; Гілки дерева закривають вікна; Потрібна обрізка дерев; Суха гілка впала на машину; Заросли кущі перекривають тротуар; Газон не косять вже місяць; Трава по пояс у дворі; Аварійне дерево біля дитячого майданчика; Гілля падає на голови людям; Дерево нахилилось і може впасти; Потрібно прибрати поламані гілки після бурі",
	},
	{
		Name:        "Прибирання вулиць",
		Description: "Прибирання сміття, снігу, листя, посипання доріг, робота двірників",
		Keywords:    "Вулицю не прибирають після снігу; Дуже слизько на тротуарі; Купи листя не вивезли; Двір брудний і не прибраний; Сніг не чистять біля будинку; Ожеледиця на тротуарі люди падають; Пісок на дорогу не сипали; Двірник не виходить на роботу; Бруд на тротуарі після ремонту; Калюжі стоять не відводиться вода; Слизько біля магазину; Сходи не почистили від снігу; Потрібно прибрати сміття у дворі; Не посипають вулицю в ожеледицю; Листя гниє ніхто не прибирає",
	},
	{
		Name:        "Вивезення ТПВ",
		Description: "Вивіз сміття, контейнери, сміттєвози, графік вивозу, сортування",
		Keywords:    "Не вивозять сміття біля будинку; Переповнений сміттєвий бак; Порушений графік вивозу сміття; Сміттєвий майданчик у поганому стані; Сміття не вивозять вже тиждень; Контейнери переповнені все воняє; Сміттєвоз не приїжджає за графіком; Немає контейнера для сортування; Бак для сміття зламаний; Стоїть неприємний запах від сміття; Потрібен додатковий контейнер; Сміття розкидане по двору; Бездомні розкидають сміття з баків; Кришка контейнера зламана; Не вивозять великогабаритне сміття",
	},
	{
		Name:        "Громадський транспорт",
		Description: "Автобуси, тролейбуси, трамваї, маршрутки, графік руху, зупинки",
		Keywords:    "Автобус не приїжджає за розкладом; Переповнений громадський транспорт; Тролейбус довго не зʼявляється; Водій маршрутки хамить пасажирам; Чекаємо автобус по 40 хвилин; Маршрут 25 не їздить за розкладом; Скасували маршрут без попередження; Водій проїхав зупинку; Трамвай зламався на маршруті; У тролейбусі не працює опалення; Кондуктор не дає квиток; Зупинка громадського транспорту зламана; Розбите скло на зупинці; Немає інформації про розклад; Транспорт постійно запізнюється",
	},
	{
		Name:        "Муніципальна варта",
		Description: "Порушення порядку, шум, хуліганство, охорона громадського порядку",
		Keywords:    "Гучна музика після 22 години; Сусіди шумлять вночі; Порушення тиші у дворі; Бійка під будинком; Пʼяна компанія шумить біля підʼїзду; Шум від ремонту в неробочий час; Гучні крики на вулиці вночі; Молодь шумить на дитячому майданчику; Сусід слухає гучну музику щодня; Порушення громадського порядку; Хуліганство у дворі; Розбивають пляшки під вікнами; Запалили багаття у дворі; Шумне застілля заважає спати; Крики та шум від сусідів",
	},
}

func main() {
	// MongoDB connection
	mongoURI := getEnv("MONGODB_URI", "mongodb://admin:admin@localhost:27017")
	mongoDB := getEnv("MONGODB_DATABASE", "citizen_appeals_ml")
	mongoAuthSource := getEnv("MONGODB_AUTH_SOURCE", "admin")

	// Add authSource to URI if not present
	// MongoDB URI format: mongodb://[username:password@]host[:port]/[database][?options]
	// Query parameters require / before ? if no database path exists
	if !strings.Contains(mongoURI, "authSource") {
		if strings.Contains(mongoURI, "?") {
			// Query already exists, just add authSource
			mongoURI += "&authSource=" + mongoAuthSource
		} else {
			// No query yet - check if URI ends with port (no / after port)
			// Simple check: if URI doesn't have / after the last :, add / before ?
			lastColon := strings.LastIndex(mongoURI, ":")
			hasSlashAfterPort := lastColon != -1 && strings.Contains(mongoURI[lastColon:], "/")

			if !hasSlashAfterPort {
				// URI ends with :port, add / before query
				mongoURI += "/?authSource=" + mongoAuthSource
			} else {
				// URI has / after port (has database), just add ?
				mongoURI += "?authSource=" + mongoAuthSource
			}
		}
	}

	log.Println("Connecting to MongoDB...")
	log.Printf("MongoDB URI: %s", mongoURI)
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Test MongoDB connection
	err = mongoClient.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	mongoDBClient := mongoClient.Database(mongoDB)
	collection := mongoDBClient.Collection("service_embeddings")

	log.Println("Creating indexes...")
	createIndexes(collection)

	// Get service IDs from PostgreSQL (optional, but recommended)
	serviceIDs := make(map[string]int64)
	if pgConnStr := getPostgresConnectionString(); pgConnStr != "" {
		log.Println("Fetching service IDs from PostgreSQL...")
		ids, err := fetchServiceIDsFromPostgres(pgConnStr)
		if err != nil {
			log.Printf("Warning: Failed to fetch service IDs from PostgreSQL: %v", err)
			log.Println("Will use sequential IDs instead")
		} else {
			serviceIDs = ids
		}
	}

	log.Printf("Seeding %d services to MongoDB...", len(servicesData))

	inserted := 0
	updated := 0
	skipped := 0

	for i, serviceData := range servicesData {
		// Get service ID from PostgreSQL or use sequential
		serviceID := int64(i + 1)
		if id, ok := serviceIDs[serviceData.Name]; ok {
			serviceID = id
		}

		// Split keywords by semicolon and clean up
		examples := []string{}
		if serviceData.Keywords != "" {
			parts := strings.Split(serviceData.Keywords, ";")
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					examples = append(examples, trimmed)
				}
			}
		}

		// If no examples from keywords, use description as single example
		if len(examples) == 0 && serviceData.Description != "" {
			examples = []string{serviceData.Description}
		}

		embedding := ServiceEmbedding{
			ServiceID:   serviceID,
			ServiceName: serviceData.Name,
			Description: serviceData.Description,
			Examples:    examples,
			Version:     1,
			UpdatedAt:   time.Now(),
			CreatedAt:   time.Now(),
		}

		// Check if document already exists
		filter := bson.M{"service_id": serviceID}
		var existing ServiceEmbedding
		err := collection.FindOne(context.Background(), filter).Decode(&existing)

		if err == mongo.ErrNoDocuments {
			// Insert new document
			_, err = collection.InsertOne(context.Background(), embedding)
			if err != nil {
				log.Printf("Failed to insert service %s: %v", serviceData.Name, err)
				skipped++
				continue
			}
			log.Printf("✓ Inserted: %s (ID: %d, %d examples)", serviceData.Name, serviceID, len(examples))
			inserted++
		} else if err != nil {
			log.Printf("Error checking existing document for %s: %v", serviceData.Name, err)
			skipped++
			continue
		} else {
			// Update existing document (increment version)
			embedding.Version = existing.Version + 1
			embedding.CreatedAt = existing.CreatedAt // Keep original creation time
			update := bson.M{
				"$set": bson.M{
					"service_name": embedding.ServiceName,
					"description":  embedding.Description,
					"examples":     embedding.Examples,
					"version":      embedding.Version,
					"updated_at":   embedding.UpdatedAt,
				},
			}
			_, err = collection.UpdateOne(context.Background(), filter, update)
			if err != nil {
				log.Printf("Failed to update service %s: %v", serviceData.Name, err)
				skipped++
				continue
			}
			log.Printf("✓ Updated: %s (ID: %d, version %d, %d examples)", serviceData.Name, serviceID, embedding.Version, len(examples))
			updated++
		}
	}

	log.Println("")
	log.Println("Seeding completed!")
	log.Printf("  Inserted: %d", inserted)
	log.Printf("  Updated:  %d", updated)
	log.Printf("  Skipped:  %d", skipped)
	log.Printf("  Total:    %d", len(servicesData))
}

func getPostgresConnectionString() string {
	pgHost := getEnv("DB_HOST", "")
	pgPort := getEnv("DB_PORT", "5432")
	pgUser := getEnv("DB_USER", "postgres")
	pgPassword := getEnv("DB_PASSWORD", "postgres")
	pgDB := getEnv("DB_NAME", "citizen_appeals")

	// Only return connection string if DB_HOST is set
	if pgHost == "" {
		return ""
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", pgUser, pgPassword, pgHost, pgPort, pgDB)
}

func fetchServiceIDsFromPostgres(connStr string) (map[string]int64, error) {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	defer pool.Close()

	query := `SELECT id, name FROM services ORDER BY id`
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make(map[string]int64)
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		ids[name] = id
	}

	return ids, rows.Err()
}

func createIndexes(collection *mongo.Collection) {
	ctx := context.Background()

	// Index on service_id (unique)
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "service_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Printf("Warning: Failed to create service_id index: %v", err)
	}

	// Index on service_name
	indexModel = mongo.IndexModel{
		Keys: bson.D{{Key: "service_name", Value: 1}},
	}
	_, err = collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Printf("Warning: Failed to create service_name index: %v", err)
	}

	// Index on version
	indexModel = mongo.IndexModel{
		Keys: bson.D{{Key: "version", Value: 1}},
	}
	_, err = collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Printf("Warning: Failed to create version index: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
