package backend_process

const BATCH_SIZE = 20000

type TraceIdBatch struct {
	batchPos int
	processCount int
	traceIdList []string
}
