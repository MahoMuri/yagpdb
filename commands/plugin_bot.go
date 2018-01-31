package commands

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/jonas747/dcmd"
	"github.com/jonas747/discordgo"
	"github.com/jonas747/yagpdb/bot"
	"github.com/jonas747/yagpdb/bot/eventsystem"
	"github.com/jonas747/yagpdb/common"
)

var (
	CommandSystem *dcmd.System
)

func (p *Plugin) InitBot() {
	CommandSystem = dcmd.NewStandardSystem("")
	CommandSystem.Prefix = p
	CommandSystem.State = bot.State
	CommandSystem.Root.RunInDM = true

	CommandSystem.Root.AddCommand(cmdHelp, cmdHelp.GetTrigger())

	eventsystem.AddHandler(bot.RedisWrapper(HandleGuildCreate), eventsystem.EventGuildCreate)
	eventsystem.AddHandler(handleMsgCreate, eventsystem.EventMessageCreate)
}

func AddRootCommands(cmds ...*YAGCommand) {
	for _, v := range cmds {
		CommandSystem.Root.AddCommand(v, v.GetTrigger())
	}
}

func handleMsgCreate(evt *eventsystem.EventData) {
	CommandSystem.HandleMessageCreate(common.BotSession, evt.MessageCreate)
}

func (p *Plugin) Prefix(data *dcmd.Data) string {
	client, err := common.RedisPool.Get()
	if err != nil {
		log.WithError(err).Error("Failed retrieving redis connection from pool")
		return ""
	}
	defer common.RedisPool.Put(client)

	prefix, err := GetCommandPrefix(client, data.CS.ID())
	if err != nil {
		log.WithError(err).Error("Failed retrieving commands prefix")
	}

	return prefix
}

var cmdHelp = &YAGCommand{
	Name:        "Help",
	Description: "Shows help abut all or one specific command",
	CmdCategory: CategoryGeneral,
	RunInDM:     true,
	Arguments: []*dcmd.ArgDef{
		&dcmd.ArgDef{Name: "command", Type: dcmd.String},
	},

	RunFunc:  cmdFuncHelp,
	Cooldown: 10,
}

func CmdNotFound(search string) string {
	return fmt.Sprintf("Couldn't find command %q", search)
}

func cmdFuncHelp(data *dcmd.Data) (interface{}, error) {
	target := ""
	if data.Args[0] != nil {
		target = data.Args[0].Str()
	}

	var resp []*discordgo.MessageEmbed
	if target != "" {
		// Send the targetted help in the channel it was requested in
		resp = dcmd.GenerateTargettedHelp(target, data, data.ContainerChain[0], &dcmd.StdHelpFormatter{})
		if len(resp) < 1 {
			return CmdNotFound(target), nil
		}

		return resp, nil
	} else {
		// Send full help in DM
		channel, err := bot.GetCreatePrivateChannel(data.Msg.Author.ID)
		if err != nil {
			return "Something went wrong", err
		}

		resp = dcmd.GenerateHelp(data, data.ContainerChain[0], &dcmd.StdHelpFormatter{})
		for _, v := range resp {
			common.BotSession.ChannelMessageSendEmbed(channel.ID, v)
		}
	}

	return nil, nil

}

func HandleGuildCreate(evt *eventsystem.EventData) {
	client := bot.ContextRedis(evt.Context())
	g := evt.GuildCreate
	prefixExists, err := common.RedisBool(client.Cmd("EXISTS", "command_prefix:"+g.ID))
	if err != nil {
		log.WithError(err).Error("Failed checking if prefix exists")
		return
	}

	if !prefixExists {
		client.Cmd("SET", "command_prefix:"+g.ID, "-")
		log.WithField("guild", g.ID).WithField("g_name", g.Name).Info("Set command prefix to default (-)")
	}
}
