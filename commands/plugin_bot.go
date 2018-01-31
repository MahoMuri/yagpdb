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
	CommandSystem := dcmd.NewStandardSystem("")
	CommandSystem.Prefix = p

	// CommandSystem = commandsystem.NewSystem(nil, "")
	// CommandSystem.SendError = false
	// CommandSystem.CensorError = CensorError
	// CommandSystem.State = bot.State

	// CommandSystem.DefaultDMHandler = &commandsystem.Command{
	// 	Run: func(data *commandsystem.ExecData) (interface{}, error) {
	// 		return "Unknown command, only a subset of commands are available in dms.", nil
	// 	},
	// }

	// CommandSystem.Prefix = p
	// CommandSystem.RegisterCommands(cmdHelp)

	eventsystem.AddHandler(bot.RedisWrapper(HandleGuildCreate), eventsystem.EventGuildCreate)
	eventsystem.AddHandler(handleMsgCreate, eventsystem.EventMessageCreate)
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

func GenerateHelp(target string) string {
	// 	if target != "" {
	// 		return CommandSystem.GenerateHelp(target, 100)
	// 	}

	// 	categories := make(map[CommandCategory][]*CustomCommand)

	// 	for _, v := range CommandSystem.Commands {
	// 		cast := v.(*CustomCommand)
	// 		categories[cast.Category] = append(categories[cast.Category], cast)
	// 	}

	// 	out := "```ini\n"

	// 	out += `[Legend]
	// #
	// #Command   = {alias1, alias2...} <required arg> (optional arg) : Description
	// #
	// #Example:
	// Help        = {hlp}   (command)       : blablabla
	// # |             |          |                |
	// #Comand name, Aliases,  optional arg,    Description

	// `

	// 	// Do it manually to preserve order
	// 	out += "[General] # General YAGPDB commands"
	// 	out += generateComandsHelp(categories[CategoryGeneral]) + "\n"

	// 	out += "\n[Tools]"
	// 	out += generateComandsHelp(categories[CategoryTool]) + "\n"

	// 	out += "\n[Moderation] # These are off by default"
	// 	out += generateComandsHelp(categories[CategoryModeration]) + "\n"

	// 	out += "\n[Misc/Fun] # Fun commands for family and friends!"
	// 	out += generateComandsHelp(categories[CategoryFun]) + "\n"

	// 	out += "\n[Debug/Maintenance] # Commands for maintenance and debug mainly."
	// 	out += generateComandsHelp(categories[CategoryDebug]) + "\n"

	// 	unknown, ok := categories[CommandCategory("")]
	// 	if ok && len(unknown) > 1 {
	// 		out += "\n[Unknown] # ??"
	// 		out += generateComandsHelp(unknown) + "\n"
	// 	}

	// 	out += "```"
	return ""
}

// func generateComandsHelp(cmds []*CustomCommand) string {
// 	out := ""
// 	for _, v := range cmds {
// 		if !v.HideFromHelp {
// 			out += "\n" + v.GenerateHelp("", 100, 0)
// 		}
// 	}
// 	return out
// }

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

	// Fetch the prefix if ther command was not run in a dm
	// footer := ""
	// if data.Source != dcmd.DMSource && target == "" {
	// 	prefix, err := GetCommandPrefix(data.Context().Value(CtxKeyRedisClient).(*redis.Client), data.GS.ID())
	// 	if err != nil {
	// 		return "Error communicating with redis", err
	// 	}

	// 	footer = "**No command prefix set, you can still use commands through mentioning the bot\n**"
	// 	if prefix != "" {
	// 		footer = fmt.Sprintf("**Command prefix: %q**\n", prefix)
	// 	}
	// }

	// if target == "" {
	// 	footer += "**Support server:** https://discord.gg/0vYlUK2XBKldPSMY\n**Control Panel:** https://yagpdb.xyz/manage\n"
	// }

	// channelId := data.Msg.ChannelID

	// help := GenerateHelp(target)
	// if target == "" && data.Source != commandsystem.SourceDM {
	// 	privateChannel, err := bot.GetCreatePrivateChannel(data.Message.Author.ID)
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	channelId = privateChannel.ID
	// }

	// if help == "" {
	// 	help = "Command not found"
	// }

	// dutil.SplitSendMessagePS(common.BotSession, channelId, help+"\n"+footer, "```ini\n", "```", false, false)
	// if data.Source != commandsystem.SourceDM && target == "" {
	// 	return "You've Got Mail!", nil
	// } else {
	// 	return "", nil
	// }
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
