package main

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

type BotCommand int

const (
	bcUnknown BotCommand = iota
	bcStart
	bcHelp
	bcInfo
	bcSerAdd
	bcSerEdit
	bcSerDel
	bcMyId
	bcWl
	bcWlAdd
	bcWlDel
	bcStatus
	bcBook
	bcRelease
)
const timeFormat = "2006-01-02 15:04:05"
const wrongCommand = "Неверная команда"

var BotCommands = [bcRelease + 1]string{"", "/start", "/help", "/info", "/seradd", "/seredit", "/serdel", "/myid", "/wl", "/wladd", "/wldel", "/status", "/book", "/release"}

var bookRegexp = regexp.MustCompile(`(?i)\/book(?:@.*bot)?\s+(?P<num>\d+)\s+(?P<tl>\d+)(?P<tp>m|h|d)(?:\s+(?P<desc>.*))?`)  // /book@bot 2 30m TASK-123
var releaseRegexp = regexp.MustCompile(`(?i)\/release(?:@.*bot)?(?:\s+(?P<num>\d+))?`)                                      // /release@bot 1
var serAddRegexp = regexp.MustCompile(`(?i)\/seradd(?:@.*bot)?\s+{(?P<name>[^}]+)}(?:\s+{(?P<desc>.+)})?`)                  // /seradd {server_name} {desc ription}
var serEditRegexp = regexp.MustCompile(`(?i)\/seredit(?:@.*bot)?\s+(?P<num>\d+)\s+{(?P<name>[^}]+)}(?:\s+{(?P<desc>.+)})?`) // /seredit 123 {server_name} {desc ription}
var serDelRegexp = regexp.MustCompile(`(?i)\/serdel(?:@.*bot)?\s+(?P<name>.+)`)                                             // /serdel@bot server_name
var wlAddRegexp = regexp.MustCompile(`(?i)\/wladd(?:@.*bot)?\s+(?P<num>\d+)(?:\s+(?P<desc>.*))?`)                           // /wladd@bot 123
var wlDelRegexp = regexp.MustCompile(`(?i)\/wldel(?:@.*bot)?\s+(?P<num>\d+)`)                                               // /wldel 123

var bookCommand = BotCommands[bcBook] + " <n> <t> <d> - бронирование сервера под номером n на время t (30m - 30 минут, 1h - 1 час, 2d - 2 дня) с описанием d (не обязательно)\n    "

//-----------------------------------------------------------

func FindBotCommand(command string) BotCommand {
	splt := strings.Split(command, "@") // /cmnd@botname
	command = strings.ToLower(splt[0])
	splt = strings.Split(command, " ") // /cmnd params
	command = strings.ToLower(splt[0])

	for k, v := range BotCommands {
		if command == v {
			return BotCommand(k)
		}
	}

	return bcUnknown
}

func cmnd_Start(userId int) string {
	if settings.AdminId == 0 {
		settings.AdminId = userId
		UpdateSettings()
	}

	return cmnd_Help(userId)
}

func cmnd_Help(userId int) string {
	msg := "Бронирование серверов\n\n" +
		"Команды:\n    " +
		BotCommands[bcHelp] + " - справка (этот текст)\n    " +
		BotCommands[bcInfo] + " - информация о серверах\n    " +
		BotCommands[bcStatus] + " - статус занятости серверов\n    " +
		bookCommand +
		BotCommands[bcRelease] + " - освободить занятый сервер\n    " +
		BotCommands[bcMyId] + " - узнать свой id (для администратора)"

	if settings.AdminId == userId {
		return msg +
			"\n\nВы являетесь администратором бота\n" +
			"Команды, доступные только администратору:\n    " +
			BotCommands[bcSerAdd] + " <{name}> <{desc}> - добавить сервер с именем {name} (в фигурных скобках) и описанием {desc} (в фигурных скобках, не обязательно)\n    " +
			BotCommands[bcSerEdit] + " <num> <{name}> <{desc}> - изменить у сервера под номером num имя {name} (в фигурных скобках) и описание {desc} (в фигурных скобках, не обязательно)\n    " +
			BotCommands[bcSerDel] + " <name> - удалить сервер с именем name\n    " +
			BotCommands[bcWl] + " - посмотреть белый список пользователей\n    " +
			BotCommands[bcWlAdd] + " <userId> - добавить пользователя в белый список\n    " +
			BotCommands[bcWlDel] + " <userId> - удалить пользователя из белого списка\n    "
	}

	return msg
}

func cmnd_Info(userId int) string {
	if !UserInWhiteList(userId) {
		return ""
	}

	if len(settings.Servers) == 0 {
		return "Список серверов пуст"
	}

	builder := strings.Builder{}
	for i, v := range settings.Servers {
		builder.WriteString(strconv.Itoa(i+1) + ". " + v.Name + "\n")
		if v.Info != "" {
			builder.WriteString("     " + v.Info + "\n")
		}
	}

	return builder.String()
}

func cmnd_Status(userId int) string {
	if !UserInWhiteList(userId) {
		return ""
	}

	if len(settings.Servers) == 0 {
		return "Список серверов пуст"
	}

	builder := strings.Builder{}
	for i, v := range settings.Servers {
		if v.IsFree {
			builder.WriteString(strconv.Itoa(i+1) + ". " + v.Name + " - свободен\n")
		} else {
			builder.WriteString(strconv.Itoa(i+1) + ". " + v.Name + " - занят\n")
			builder.WriteString(
				"     @" + v.ByName + " с " + time.Unix(v.From, 0).Format(timeFormat) +
					" до " + time.Unix(v.To, 0).Format(timeFormat) + "\n")
		}
	}

	return builder.String()
}

func cmnd_Book(command string, userId int, userName string) string {
	wrongCommand := "Неверная команда\n" + "Правильный формат:\n" + bookCommand

	if !UserInWhiteList(userId) {
		return ""
	}

	if len(settings.Servers) == 0 {
		return "Список серверов пуст"
	}

	for _, v := range settings.Servers {
		if v.ById == userId {
			return "У вас уже есть забронировнный сервер " + v.Name
		}
	}

	if !bookRegexp.MatchString(command) {
		return wrongCommand
	}
	res := bookRegexp.FindStringSubmatch(command)
	l := len(res)
	if l < 4 || l > 5 {
		return wrongCommand
	}
	num, err := strconv.Atoi(res[1])
	if err != nil {
		return wrongCommand
	}
	num -= 1
	tli, err := strconv.Atoi(res[2])
	if err != nil {
		return wrongCommand
	}
	tl := int64(tli)
	tp := strings.ToLower(res[3])
	desc := ""
	if l == 4 {
		desc = res[4]
	}

	if num > len(settings.Servers)-1 {
		return "Сервера под таким номером не существует"
	}

	if !settings.Servers[num].IsFree {
		return "Сервер " + settings.Servers[num].Name + " уже забронирован @" + settings.Servers[num].ByName +
			" с " + time.Unix(settings.Servers[num].From, 0).Format(timeFormat) +
			" до " + time.Unix(settings.Servers[num].To, 0).Format(timeFormat)
	}

	var timePeriod int64
	switch tp {
	case "m":
		timePeriod = tl * 60
	case "h":
		timePeriod = tl * 60 * 60
	case "d":
		timePeriod = tl * 60 * 60 * 24
	}

	settings.Servers[num].IsFree = false
	settings.Servers[num].From = time.Now().Unix()
	settings.Servers[num].To = settings.Servers[num].From + timePeriod
	settings.Servers[num].ById = userId
	settings.Servers[num].ByName = userName
	settings.Servers[num].Desc = desc

	UpdateSettings()

	return "Сервер " + settings.Servers[num].Name + " забронирован до " +
		time.Unix(settings.Servers[num].To, 0).Format(timeFormat) + "\n" + BotCommands[bcStatus]
}

func cmnd_Release(command string, userId int) string {
	checkUser := true
	num := -1

	if !UserInWhiteList(userId) {
		return ""
	}

	if !releaseRegexp.MatchString(command) {
		return wrongCommand
	}
	res := bookRegexp.FindStringSubmatch(command)
	if len(res) == 2 {
		n, err := strconv.Atoi(res[1])
		if err != nil {
			return wrongCommand
		}
		if settings.AdminId != userId {
			return wrongCommand
		}
		num = n
		checkUser = false
	}

	if checkUser {
		for i, v := range settings.Servers {
			if v.ById == userId {
				num = i
			}
		}
		if num == -1 {
			return "У вас не было забронированных серверов"
		}
	}

	ReleaseServer(num)
	UpdateSettings()

	return "Сервер " + settings.Servers[num].Name + " свободен\n" + BotCommands[bcStatus]
}

func cmnd_MyId(userId int) string {
	return strconv.Itoa(userId)
}

func cmnd_Wl(userId int) string {
	if settings.AdminId != userId {
		return ""
	}

	if len(settings.WhiteList) == 0 {
		return "Белый список пуст"
	}

	builder := strings.Builder{}
	for _, v := range settings.WhiteList {
		builder.WriteString(strconv.Itoa(v.Id) + " - " + v.Desc + "\n")
	}

	return builder.String()
}

func cmnd_WlAdd(command string, userId int) string {
	if settings.AdminId != userId {
		return ""
	}

	if !wlAddRegexp.MatchString(command) {
		return wrongCommand
	}
	res := wlAddRegexp.FindStringSubmatch(command)
	l := len(res)
	if l < 2 || l > 3 {
		return wrongCommand
	}
	id, err := strconv.Atoi(res[1])
	if err != nil {
		return wrongCommand
	}
	desc := ""
	if l == 3 {
		desc = res[2]
	}

	for _, v := range settings.WhiteList {
		if v.Id == id {
			return "Пользователь с этим id уже добавлен"
		}
	}

	wl := WhiteList{}
	wl.Id = id
	wl.Desc = desc
	settings.WhiteList = append(settings.WhiteList, wl)
	UpdateSettings()

	return "Пользователь добавлен в белый список\n" + BotCommands[bcWl]
}

func cmnd_WlDel(command string, userId int) string {
	if settings.AdminId != userId {
		return ""
	}

	if len(settings.WhiteList) == 0 {
		return "Белый список пуст"
	}

	if !wlDelRegexp.MatchString(command) {
		return wrongCommand
	}
	res := wlDelRegexp.FindStringSubmatch(command)
	if len(res) != 2 {
		return wrongCommand
	}
	id, err := strconv.Atoi(res[1])
	if err != nil {
		return wrongCommand
	}

	num := -1
	for i, v := range settings.WhiteList {
		if v.Id == id {
			num = i
			break
		}
	}
	if num == -1 {
		return "Пользователь с указанным id не найден в белом списке"
	}

	settings.WhiteList = append(settings.WhiteList[:num], settings.WhiteList[num+1:]...)
	UpdateSettings()

	return "Пользователь удалён из белого списка\n" + BotCommands[bcWl]
}

func cmnd_SerAdd(command string, userId int) string {
	if settings.AdminId != userId {
		return ""
	}

	if !serAddRegexp.MatchString(command) {
		return wrongCommand
	}
	res := serAddRegexp.FindStringSubmatch(command)
	l := len(res)
	if l < 2 || l > 3 {
		return wrongCommand
	}
	name := res[1]
	info := ""
	if l == 3 {
		info = res[2]
	}

	for _, v := range settings.Servers {
		if v.Name == name {
			return "Сервер с этим именем уже добавлен"
		}
	}

	server := Server{}
	server.Name = name
	server.Info = info
	server.IsFree = true
	settings.Servers = append(settings.Servers, server)
	UpdateSettings()

	return "Сервер добавлен\n" + BotCommands[bcInfo]
}

func cmnd_SerEdit(command string, userId int) string {
	if settings.AdminId != userId {
		return ""
	}

	if !serEditRegexp.MatchString(command) {
		return wrongCommand
	}
	res := serEditRegexp.FindStringSubmatch(command)
	l := len(res)
	if l < 3 || l > 4 {
		return wrongCommand
	}
	id, err := strconv.Atoi(res[1])
	if err != nil {
		return wrongCommand
	}
	id -= 1
	name := res[2]
	info := ""
	if l == 4 {
		info = res[3]
	}

	for i, v := range settings.Servers {
		if i != id && v.Name == name {
			return "Сервер с этим именем уже существует"
		}
	}

	settings.Servers[id].Name = name
	settings.Servers[id].Info = info
	UpdateSettings()

	return "Сервер обновлён\n" + BotCommands[bcInfo]
}

func cmnd_SerDel(command string, userId int) string {
	if settings.AdminId != userId {
		return ""
	}

	if len(settings.Servers) == 0 {
		return "Список серверов пуск"
	}

	if !serDelRegexp.MatchString(command) {
		return wrongCommand
	}
	res := serDelRegexp.FindStringSubmatch(command)
	if len(res) != 2 {
		return wrongCommand
	}
	name := res[1]

	num := -1
	for i, v := range settings.Servers {
		if v.Name == name {
			num = i
			break
		}
	}
	if num == -1 {
		return "Сервер с указанным именем не найден"
	}

	settings.Servers = append(settings.Servers[:num], settings.Servers[num+1:]...)
	UpdateSettings()

	return "Сервер " + name + " удалён из списка\n" + BotCommands[bcInfo]
}

func UserInWhiteList(userId int) bool {
	if settings.AdminId == userId {
		return true
	}

	for _, v := range settings.WhiteList {
		if v.Id == userId {
			return true
		}
	}

	return false
}
