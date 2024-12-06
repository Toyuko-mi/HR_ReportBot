package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/telebot.v3"
)

// Структура для хранения данных пользователя
type UserState struct {
	Step    int               // Номер текущего вопроса
	Answers map[string]string // Ответы пользователя
}

func main() {
	// Настраиваем логирование в файл для проверки
	logFile, err := os.OpenFile("bot.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Ошибка открытия файла для логирования:", err)
		return
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// Токен вашего бота (замените на реальный токен)
	token := "7995352159:AAGjQfAXLUwhShaZ0DnT8-CgZ9SS8Mk7dOs" // Замените на ваш реальный токен бота

	// Целевой чат для отправки результатов (укажите правильный Chat ID)
	targetChatID := int64(-1002455432625) // Замените на реальный Chat ID вашей группы

	// Настройки для инициализации бота
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	log.Println("Бот успешно создан и запущен!")

	// Хранение состояний пользователей
	userStates := make(map[int64]*UserState)

	// Вопросы для бота
	questions := []string{
		"Постоянный или временный?",
		"ФИО сотрудника, которому нужен пропуск?",
		"Отдел, в котором работает сотрудник?",
		"Когда сотрудник начал у нас работать?",
		"Причина получения пропуска: Новый / Утерянный / Восстановление?",
	}

	// Обработчик команды /start
	bot.Handle("/start", func(c telebot.Context) error {
		chatID := c.Chat().ID
		log.Printf("Получена команда /start от пользователя %d\n", chatID)

		userStates[chatID] = &UserState{
			Step:    0,
			Answers: make(map[string]string),
		}

		// Отправляем первый вопрос
		question := questions[0]
		if err := c.Send(question); err != nil {
			log.Printf("Ошибка отправки вопроса пользователю %d: %v\n", chatID, err)
			return err
		}
		log.Printf("Первый вопрос отправлен пользователю %d: %s\n", chatID, question)

		return nil
	})

	// Обработчик текстовых сообщений для вопросов
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		chatID := c.Chat().ID
		state, exists := userStates[chatID]

		if !exists {
			log.Printf("Пользователь %d не найден в состоянии, запрашивается команда /start\n", chatID)
			return c.Send("Напиши /start, чтобы начать!")
		}

		// Сохранение ответа пользователя
		switch state.Step {
		case 0:
			state.Answers["type"] = c.Text()
		case 1:
			state.Answers["name"] = c.Text()
		case 2:
			state.Answers["department"] = c.Text()
		case 3:
			state.Answers["date"] = c.Text()
		case 4:
			state.Answers["reason"] = c.Text()
		}

		log.Printf("Ответ пользователя сохранен. Текущий шаг: %d, Пользователь %d\n", state.Step, chatID)

		// Переход к следующему шагу
		state.Step++

		// Если все вопросы заданы, отправляем результат в группу
		if state.Step >= len(questions) {
			result := generateTemplate(state.Answers)
			delete(userStates, chatID) // Удаляем состояние пользователя

			// Отправляем результат в целевой чат (группу)
			recipient := &telebot.Chat{ID: targetChatID}
			log.Printf("Пытаемся отправить результат в группу с ID %d\n", targetChatID)

			if _, err := bot.Send(recipient, result); err != nil {
				log.Printf("Ошибка отправки результата в группу с ID %d: %v\n", targetChatID, err)
				return err
			}
			log.Printf("Результат успешно отправлен в группу с ID %d\n", targetChatID)

			return nil
		}

		// Задаем следующий вопрос
		nextQuestion := questions[state.Step]
		if err := c.Send(nextQuestion); err != nil {
			log.Printf("Ошибка отправки вопроса пользователю %d: %v\n", chatID, err)
			return err
		}
		log.Printf("Следующий вопрос отправлен пользователю %d: %s\n", chatID, nextQuestion)

		return nil
	})

	log.Println("Бот запущен и готов принимать команды...")
	bot.Start()
}

// Генерация текста из шаблона
func generateTemplate(answers map[string]string) string {
	return fmt.Sprintf(
		"Тип пропуска: %s\nФИО сотрудника: %s\nОтдел: %s\nДата выхода сотрудника: %s\nПричина получения пропуска: %s\n\nДополнительные действия:\n- Подписанное NDA: Да\n- Подписанная Материальная ответственность: Да\n- Заполнен в документе численность: Да\n- Заполнена форма СБ Стажеры/кандидаты: Да",
		answers["type"],
		answers["name"],
		answers["department"],
		answers["date"],
		answers["reason"],
	)
}
