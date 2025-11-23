package aggregates

import "github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"

type Chat struct {
	*entities.Chat

	Couples []CommandResultCouple
}

type CommandResultCouple struct {
	Command      *entities.Command
	CommandMedia []*entities.CommandMedia

	// Can be nil, if command is not executed yet.
	Result      *entities.Result
	ResultMedia []*entities.ResultMedia

	Model *entities.Model
	Tool  *entities.Tool
}

func NewChat(chat *entities.Chat, couples []CommandResultCouple) (*Chat, error) {
	err := chat.Validate()
	if err != nil {
		return nil, err
	}

	for _, couple := range couples {
		err = couple.Command.Validate()
		if err != nil {
			return nil, err
		}

		if couple.Result != nil {
			err = couple.Result.Validate()
			if err != nil {
				return nil, err
			}
		}
	}

	return &Chat{Chat: chat, Couples: couples}, nil
}

func (c *Chat) SetTitle(title string) error {
	oldTitle := c.Title
	c.Title = title

	err := c.Chat.Validate()
	if err != nil {
		c.Title = oldTitle

		return err
	}

	return nil
}
