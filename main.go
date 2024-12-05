package main

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/telebot.v3"
)

// Структура для хранения данных пользователя
type UserState struct {
	Step    int               // Номер текущего вопроса
	Answers map[string]string // Ответы пользователя
}

// Шаблон текста для итогового сообщения
const templateText = `
Привет, {{name}}!
Ты живёшь в {{city}} и любишь заниматься {{hobby}}.
Спасибо за ответы!
`

func main() {
	token := "7995352159:AAGjQfAXLUwhShaZ0DnT8-CgZ9SS8Mk7dOs"

	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	// Хранение состояний пользователей
	userStates := make(map[int64]*UserState)

	// Вопросы для бота
	questions := []string{
		"Какой пропуск? Постоянный пропуск / Врменный пропуск",
		"ФИО сотрудника, которому нужен пропуск?",
		"Отдел, в котором работает сотрудник?",
		"Когда сотрудник начал у нас работать?",
		"Получение пропуска: Нового / Утерянного / Восстановление ?",
	}

	// Обработчик команды /start
	bot.Handle("/start", func(c telebot.Context) error {
		chatID := c.Chat().ID
		userStates[chatID] = &UserState{
			Step:    0,
			Answers: make(map[string]string),
		}
		return c.Send("Привет! Чтобы выдать сотруднику пропуск, нам нужна информация. Ответь на вопросы\n" + questions[0])
	})

	// Обработчик входящих сообщений
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		chatID := c.Chat().ID
		state, exists := userStates[chatID]

		if !exists {
			return c.Send("Напиши /start, чтобы начать!")
		}

		// Сохранение ответа
		if state.Step == 0 {
			state.Answers["type"] = c.Text()
		} else if state.Step == 1 {
			state.Answers["name"] = c.Text()
		} else if state.Step == 2 {
			state.Answers["department"] = c.Text()
		} else if state.Step == 3 {
			state.Answers["date"] = c.Text()
		} else if state.Step == 4 {
			state.Answers["reason"] = c.Text()
		}

		// Следующий шаг
		state.Step++

		// Если вопросы закончились, отправляем результат
		if state.Step >= len(questions) {
			result := generateTemplate(state.Answers)
			delete(userStates, chatID) // Удаляем состояние пользователя
			return c.Send(result)
		}

		// Задаём следующий вопрос
		return c.Send(questions[state.Step])
	})

	log.Println("Бот запущен!")
	bot.Start()
}

// Генерация текста из шаблона
func generateTemplate(answers map[string]string) string {
	return fmt.Sprintf(
		"%s \nФИО сотрудника: %s\nОтдел: %s \nРуководитель: Москаленко \nДата выхода сотрудника: %s \n\nПолучение: %s \nПодписанное NDA: Да \nПодписанная Мат. ответственность: Да \nЗаполнен в документе Численность: Да \nЗаполнена форма СБ Стажеры/кандидаты: Да",
		answers["type"],
		answers["name"],
		answers["department"],
		answers["date"],
		answers["reason"],
	)
}
