package serverapp

import (
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os/signal"
	"passkeeper/internal/config"
	"passkeeper/internal/database"
	"passkeeper/internal/logger"
	"passkeeper/internal/router"
	"passkeeper/internal/server"
	"syscall"
)

// New производим старт приложения
func New(cnf *config.CliConfig) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT) // Контекст для правильной остановки синхронизации
	defer func() {
		logger.Log.Info("Cancel context")
		cancel()
	}()

	pool, err := database.NewDB(cnf.DatabaseDSN, 5, 5)
	// Инициализируем базу данных
	if err != nil {
		return err
	}
	// Вызываем функцию закрытия базы данных
	defer pool.Close()
	// Производим миграции базы
	if err = pool.Migrate(); err != nil {
		return err
	}

	wg := new(errgroup.Group)
	serv := server.NewServer(ctx, router.NewRouter(pool, cnf), cnf.Address)
	// Запускаем сервер
	wg.Go(func() error {
		sErr := serv.S.ListenAndServe()
		if sErr != nil && !errors.Is(sErr, http.ErrServerClosed) {
			return sErr
		}
		return nil
	})
	// Регистрируем прослушиватель для завершения сервера
	<-ctx.Done()
	logger.Log.Info("Stopping server")
	serv.Close()

	// Ожидаем завершения всех горутин перед завершением программы
	if err = wg.Wait(); err != nil {
		logger.Log.Error(err)
	}
	logger.Log.Info("End Server")
	return nil
}
