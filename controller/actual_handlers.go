package controller

import (
	"bakalover/hikari-bot/game"
	"bakalover/hikari-bot/util"
	"errors"
	"fmt"
	"log"
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
			fmt.Sprintf("%s %v %s", c.TeleCtx.Chat().FirstName, c.CTK.ThreadId, c.TeleCtx.Sender().FirstName))
	}

	c.Game = game.NewGame(c.CTK, c.DbConn, c.Dicts)
	cont, initKana, err := c.Game.Continue()
	if err != nil {
		return fmt.Errorf("start game handler error: %w", err)
	}

	if cont {
		util.Reply(c.TeleCtx, fmt.Sprintf("%s\nПоследняя кана: 「%s」", ContinueGame, initKana))
	} else {
		initKana, err := c.Game.StartGame()
		if err != nil {
			return errors.New("game start handler error:\n" + err.Error())
		}

		msg := fmt.Sprintf("%s\nПервая кана: 「%s」", Greetings, initKana)
		util.Reply(c.TeleCtx, msg)
	}

	c.StopTTL()
	return nil
}

type StopGameHandler struct {
}

func (h *StopGameHandler) Handle(c *WorkerContext) error {
	if c.Game == nil {
		util.Reply(c.TeleCtx, IsNotStartedError)
		return errors.New("stop game handler error: is not started " +
			fmt.Sprintf("%s %v %s", c.TeleCtx.Chat().FirstName, c.CTK.ThreadId, c.TeleCtx.Sender().FirstName))
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
	c.RefreshTTL()

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
		msg = fmt.Sprintf(WrongOrder, c.TeleCtx.Sender().FirstName)
	case game.WordNotJapanese:
		// Just Ignoring
		break
	case game.DictsNotAnswering:
		msg = DictsUnavailable
	case game.NoSpeachPart:
		msg = MustBeNoun
	case game.NoSuchWord:
		msg = DidntFindWord
	case game.GotDoubledWord:
		msg = SameWord
	case game.CantJoinWords:
		msg = CantJoinWord

	// Difficult cases
	case game.GotError:
		msg = CallAdmin
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
		c.RefreshTTL()

		return nil
	}

	util.Reply(c.TeleCtx, msg)

	return nil
}
