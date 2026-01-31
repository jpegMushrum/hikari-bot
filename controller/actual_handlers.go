package controller

import (
	"bakalover/hikari-bot/game"
	"bakalover/hikari-bot/util"
	"errors"
	"fmt"
	"log"
)

const (
	HelpInfo = `
	Справка по коммандам:
	/help - Справка по командам
	/rules - Правила игры
	/start_game - Начать игру
	/stop_game - Закончить игру и вывести результаты`

	Unknown = "Неизвестная команда"
	Rules   = `
	Правила:
	1. Два или более человек по очереди играют.

	2. Допускаются только существительные.

	3. Игрок, который выбирает слово, оканчивающееся на ん, 
	проигрывает игру, поскольку японское слово не начинается с 
	этого символа.
	
	4. Слова не могут повторяться.
	Пример: 
	-> сакура	(さくら)	
	-> радио    (ラジオ) 	
	-> онигири  (おにぎり)	
	-> рису 	(りす)		
	-> сумо 	(すもう)
	-> удон 	(うどん)
	Дополнительно: для удобства можно вводить слова как в форме кандзи так и в чистой кане`

	UnknownCommand = "Неизвестная комманда"

	AlreadyRunningError = "Игра уже запущена！"

	Greetings = "Раунд начинается!"

	IsNotStartedError = "Игра ещё не началась!"
)

type HelpHandler struct {
}

func (h *HelpHandler) Handle(c *WorkerContext) error {
	util.Reply(c.TeleCtx, HelpInfo)
	return nil
}

type RulesHandler struct {
}

func (h *RulesHandler) Handle(c *WorkerContext) error {
	util.Reply(c.TeleCtx, Rules)
	return nil
}

type StartGameHandler struct {
}

func (h *StartGameHandler) Handle(c *WorkerContext) error {
	if c.Game != nil {
		util.Reply(c.TeleCtx, AlreadyRunningError)
		return errors.New("start game handler error: already started " +
			fmt.Sprintf("%s %v %s", c.TeleCtx.Chat().FirstName, c.Ctk.ThreadId, c.TeleCtx.Sender().FirstName))
	}

	c.Game = game.NewGame(c.Ctk, c.DbConn, c.Dicts)
	initKana, err := c.Game.StartGame()
	if err != nil {
		return errors.New("game start handler error:\n" + err.Error())
	}

	msg := fmt.Sprintf("%s\nПервая кана: %s", Greetings, initKana)
	util.Reply(c.TeleCtx, msg)

	return nil
}

type StopGameHandler struct {
}

func (h *StopGameHandler) Handle(c *WorkerContext) error {
	if c.Game == nil {
		util.Reply(c.TeleCtx, IsNotStartedError)
		return errors.New("stop game handler error: is not started " +
			fmt.Sprintf("%s %v %s", c.TeleCtx.Chat().FirstName, c.Ctk.ThreadId, c.TeleCtx.Sender().FirstName))
	}

	result, err := c.Game.FormStats()
	if err != nil {
		return errors.New("game stop handler error:\n" + err.Error())
	}

	err = c.Game.StopGame()
	if err != nil {
		return errors.New("game stop handler error:\n" + err.Error())
	}

	c.Game = nil

	util.Reply(c.TeleCtx, result)
	return nil
}

type NextWordGameHandler struct {
}

func (h *NextWordGameHandler) Handle(c *WorkerContext) error {
	if c.Game == nil {
		log.Println("Ignoring message: " + c.TeleCtx.Text())
		return nil
	}

	result, err := c.Game.HandleNextWord(c.TeleCtx)

	var msg string
	switch result {
	// Simple Cases
	case game.Success:
		msg = c.Game.ResultMessage
	case game.FoundLastPerson:
		msg = fmt.Sprintf("Неправильная очередь! %s добавил последнее слово", c.TeleCtx.Sender().FirstName)
	case game.WordNotJapanese:
		// Just Ignoring
		break
	case game.DictsNotAnswering:
		msg = "Словари не доступны, попробуйте чуть позже"
	case game.NoSpeachPart:
		msg = "Слово должно быть существительным!"
	case game.NoSuchWord:
		msg = "Слово не найдено!"
	case game.GotDoubledWord:
		msg = "Такое слово уже было!"
	case game.CantJoinWords:
		msg = "Это слово нельзя присоединить!"

	// Difficult cases
	case game.GotError:
		msg = "Произошла непредвиденная ошибка, позовите администратора"
		util.Reply(c.TeleCtx, msg)
		return errors.New("next word handler error:\n" + err.Error())

	case game.GotEndWord:
		stats, err := c.Game.FormStats()
		if err != nil {
			return errors.New("next word handler error:\n" + err.Error())
		}

		msg = fmt.Sprintf("%s\n%s", c.Game.ResultMessage, stats)
		util.Reply(c.TeleCtx, msg)

		err = c.Game.StopGame()
		if err != nil {
			return errors.New("next word handler error:\n" + err.Error())
		}

		c.Game = nil

		return nil
	}

	util.Reply(c.TeleCtx, msg)

	return nil
}
