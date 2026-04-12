package scheduler

import (
	"time"

	"github.com/dyallo/pricenexus/internal/db"
	"github.com/dyallo/pricenexus/internal/scraper"
	"github.com/sirupsen/logrus"
)

type Scheduler struct {
	repo     db.Repository
	interval time.Duration
	log      *logrus.Logger
	stopChan chan bool
}

func NewScheduler(repo db.Repository, intervalMinutes int, log *logrus.Logger) *Scheduler {
	return &Scheduler{
		repo:     repo,
		interval: time.Duration(intervalMinutes) * time.Minute,
		log:      log,
		stopChan: make(chan bool),
	}
}

func (s *Scheduler) Start() {
	ticker := time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.runScraping()
			case <-s.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
	s.log.Info("Scheduler started")
}

func (s *Scheduler) Stop() {
	s.stopChan <- true
	s.log.Info("Scheduler stopped")
}

func (s *Scheduler) runScraping() {
	s.log.Info("Running scheduled scraping...")

	shops, err := s.repo.GetAllShops()
	if err != nil || len(shops) == 0 {
		s.log.Warn("No shops configured")
		return
	}

	scrapers := scraper.GetAllScrapers()
	for _, scraper := range scrapers {
		s.log.Infof("Scraping %s...", scraper.Name())
		results, err := scraper.Search("ultimo producto")
		if err != nil {
			s.log.Errorf("Error scraping %s: %v", scraper.Name(), err)
			continue
		}
		s.log.Infof("Found %d results from %s", len(results), scraper.Name())
	}
}
