package core

import (
	"fmt"

	"github.com/dmsRosa6/Chirp/data_structs"
	"github.com/dmsRosa6/Chirp/grammar"
)

type IndexNode struct {
	isQueue bool // the check on is queue should happen first as a false represents a just a path
	queue   *Queue
}

type QueueManager struct {
	tree   data_structs.Trie[IndexNode]
	queues map[string]*Queue
}

func NewQueueManager() *QueueManager {
	return &QueueManager{
		tree:   *data_structs.NewTrie[IndexNode](grammar.QUEUE_PATH_DELIMITER, grammar.QUEUE_PATH_WILDCARD),
		queues: make(map[string]*Queue),
	}
}

func (qm *QueueManager) AddQueue(queue *Queue) error {
	if _, ok := qm.queues[queue.fullPath]; ok {
		return fmt.Errorf("the queue %s already exists. no operation done", queue.fullPath)
	}

	qm.queues[queue.fullPath] = queue
	qm.tree.Add(queue.fullPath, IndexNode{isQueue: true, queue: queue})
	return nil
}

func (qm *QueueManager) GetByFullPath(path string) (Queue, error) {
	if q, ok := qm.queues[path]; !ok {
		return Queue{}, fmt.Errorf("the queue %s does not exist.", path)
	} else {
		return *q, nil
	}
}

func (qm *QueueManager) ExistsByFullPath(path string) bool {
	_, ok := qm.queues[path]
	return ok
}

func (qm *QueueManager) Lookup(exp string) {

}
