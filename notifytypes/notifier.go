package notifytypes

type Notifier interface {
	Notify(changedFile, newContent string) error
}

var NotifyTypeMap = map[string]Notifier{
	"dingding": &DingDing{},
	"shell":    &Shell{},
}

func IsNotifier(t string) (Notifier, bool) {
	if v, ok := NotifyTypeMap[t]; ok {
		return v, ok
	}
	return nil, false
}
