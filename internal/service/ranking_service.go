package service

import (
	"context"
	"github.com/ecodeclub/ekit/queue"
	"github.com/gotomicro/ekit/slice"
	"log"
	"math"
	"time"
	intrv1 "webook/api/proto/gen/intr/v1"
	"webook/internal/domain"
	"webook/internal/repository"
)

type RankingService interface {
	TopN(ctx context.Context) error
}

type BatchRankingService struct {
	intrSvc   intrv1.InteractiveServiceClient
	artSvc    ArticleService
	repo      repository.RankingRepository
	batchSize int
	scoreFunc func(likeCnt int64, utime time.Time) float64
	n         int
}

func NewBatchRankingService(intrSvc intrv1.InteractiveServiceClient, artSvc ArticleService) RankingService {
	return &BatchRankingService{
		intrSvc: intrSvc,
		artSvc:  artSvc,
		scoreFunc: func(likeCnt int64, utime time.Time) float64 {
			duration := time.Since(utime).Seconds()
			return float64(likeCnt-1) / math.Pow(duration+2, 1.5)
		}}
}

func (b *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := b.topN(ctx)
	if err != nil {
		return err
	}
	log.Println(arts)
	return b.repo.ReplaceTopN(ctx, arts)
}

func (b *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	start := time.Now()
	ddl := start.Add(-7 * 24 * time.Hour)
	offset := 0
	type Score struct {
		score float64
		art   domain.Article
	}
	topN := queue.NewPriorityQueue(b.n, func(src Score, dst Score) int {
		if src.score > dst.score {
			return 1
		} else if src.score == dst.score {
			return 0
		} else {
			return -1
		}
	})
	for {
		arts, err := b.artSvc.ListPub(ctx, start, offset, b.batchSize)
		if err != nil {
			return nil, err
		}
		ids := slice.Map(arts, func(idx int, src domain.Article) int64 {
			return src.Id
		})
		intrResp, err := b.intrSvc.GetByIds(ctx, &intrv1.GetByIdsRequest{
			Biz: "article", Ids: ids,
		})
		if err != nil {
			return nil, err
		}
		intrMap := intrResp.Intrs
		for _, art := range arts {
			score := b.scoreFunc(intrMap[art.Id].LikeCnt, art.Utime)
			entry := Score{
				score: score,
				art:   art,
			}
			err := topN.Enqueue(entry)
			if err == queue.ErrOutOfCapacity {
				minEntry, _ := topN.Dequeue()
				if minEntry.score < score {
					_ = topN.Enqueue(entry)
				} else {
					_ = topN.Enqueue(minEntry)
				}
			}
		}
		offset = offset + len(arts)
		if len(arts) < b.batchSize || arts[len(arts)-1].Utime.Before(ddl) {
			break
		}
	}
	res := make([]domain.Article, b.n)
	for i := topN.Len() - 1; i >= 0; i-- {
		entry, _ := topN.Dequeue()
		res[i] = entry.art
	}
	return res, nil
}
