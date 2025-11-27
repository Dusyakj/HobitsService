package service

import (
	"context"
	"fmt"
	"time"

	"HobitsService/internal/domain"
	"HobitsService/internal/repository"
)

// ReminderService сервис для управления напоминаниями
type ReminderService struct {
	reminderRepo repository.HabitReminderRepository
	habitRepo    repository.HabitRepository
	habitService *HabitService
}

// NewReminderService создает новый ReminderService
func NewReminderService(
	reminderRepo repository.HabitReminderRepository,
	habitRepo repository.HabitRepository,
	habitService *HabitService,
) *ReminderService {
	return &ReminderService{
		reminderRepo: reminderRepo,
		habitRepo:    habitRepo,
		habitService: habitService,
	}
}

// CreateReminder создает напоминание
func (s *ReminderService) CreateReminder(ctx context.Context, habitID, userID int, reminderDate time.Time) (*domain.HabitReminder, error) {
	reminder := domain.NewHabitReminder(habitID, userID, reminderDate)
	return s.reminderRepo.CreateReminder(ctx, reminder)
}

// GenerateRemindersForToday генерирует напоминания на сегодня для пользователя
func (s *ReminderService) GenerateRemindersForToday(ctx context.Context, userID int) ([]*domain.HabitReminder, error) {
	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	// Получаем существующие напоминания на сегодня
	existingReminders, err := s.reminderRepo.GetRemindersByUserIDAndDate(ctx, userID, todayDate)
	if err != nil {
		existingReminders = []*domain.HabitReminder{}
	}

	// Создаем map существующих напоминаний для быстрой проверки
	existingHabitIDs := make(map[int]bool)
	for _, reminder := range existingReminders {
		existingHabitIDs[reminder.HabitID] = true
	}

	// Получаем все активные привычки пользователя
	habits, err := s.habitRepo.GetActiveHabitsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get habits: %w", err)
	}

	var allReminders []*domain.HabitReminder
	allReminders = append(allReminders, existingReminders...)

	for _, habit := range habits {
		// Пропускаем если напоминание уже существует
		if existingHabitIDs[habit.ID] {
			continue
		}

		// Проверяем, нужно ли подтверждение сегодня
		if s.habitService.isHabitScheduledForDate(habit, todayDate) {
			reminder := domain.NewHabitReminder(habit.ID, userID, todayDate)
			created, err := s.reminderRepo.CreateReminder(ctx, reminder)
			if err != nil {
				fmt.Printf("failed to create reminder for habit %d: %v\n", habit.ID, err)
				continue
			}
			allReminders = append(allReminders, created)
		}
	}

	return allReminders, nil
}

// GetRemindersByDate получает напоминания на дату
func (s *ReminderService) GetRemindersByDate(ctx context.Context, date time.Time) ([]*domain.HabitReminder, error) {
	return s.reminderRepo.GetRemindersByDate(ctx, date)
}

// GetRemindersByUserAndDate получает напоминания пользователя на дату
func (s *ReminderService) GetRemindersByUserAndDate(ctx context.Context, userID int, date time.Time) ([]*domain.HabitReminder, error) {
	return s.reminderRepo.GetRemindersByUserIDAndDate(ctx, userID, date)
}

// MarkReminderAsCompleted отмечает напоминание как выполненное
func (s *ReminderService) MarkReminderAsCompleted(ctx context.Context, reminderID int) (*domain.HabitReminder, error) {
	reminder, err := s.reminderRepo.GetReminderByID(ctx, reminderID)
	if err != nil {
		return nil, err
	}

	reminder.MarkAsCompleted()
	return s.reminderRepo.UpdateReminder(ctx, reminder)
}

// MarkReminderAsIncomplete отмечает напоминание как невыполненное
func (s *ReminderService) MarkReminderAsIncomplete(ctx context.Context, reminderID int) (*domain.HabitReminder, error) {
	reminder, err := s.reminderRepo.GetReminderByID(ctx, reminderID)
	if err != nil {
		return nil, err
	}

	reminder.MarkAsIncomplete()
	return s.reminderRepo.UpdateReminder(ctx, reminder)
}
