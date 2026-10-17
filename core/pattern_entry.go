package core

import (
	"sort"

	"github.com/dmsRosa6/Chirp/model"
)

type patternEntry struct {
	pattern string
	subs    map[subKey]*Subscription
	groups  map[string]*Group
}

func (e *patternEntry) snapshot() model.Subject {
	out := model.Subject{Pattern: e.pattern}

	for _, s := range e.subs {
		out.Subscribers = append(out.Subscribers, model.Subscriber{
			SessionID: s.Session.ID(),
			ID:        s.ID,
		})
	}
	sortSubscribers(out.Subscribers)

	for id, g := range e.groups {
		mg := model.Group{ID: id}
		for _, s := range g.subs {
			mg.Subscribers = append(mg.Subscribers, model.Subscriber{
				SessionID: s.Session.ID(),
				ID:        s.ID,
			})
		}
		sortSubscribers(mg.Subscribers)
		out.Groups = append(out.Groups, mg)
	}
	sort.Slice(out.Groups, func(i, j int) bool { return out.Groups[i].ID < out.Groups[j].ID })

	return out
}

func sortSubscribers(s []model.Subscriber) {
	sort.Slice(s, func(i, j int) bool {
		if s[i].SessionID != s[j].SessionID {
			return s[i].SessionID < s[j].SessionID
		}
		return s[i].ID < s[j].ID
	})
}
