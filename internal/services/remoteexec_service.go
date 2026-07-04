package services

import (
	"context"

	"network-scanner/internal/contracts"
	"network-scanner/internal/remoteexec"
)

// RemoteExecService реализация RemoteExecService
type RemoteExecService struct{}

func (s *RemoteExecService) Execute(ctx context.Context, req contracts.RemoteExecRequest) (contracts.RemoteExecResponse, error) {
	// Конвертация в internal формат
	internalReq := remoteexec.Request{
		Transport:     req.Transport,
		Target:        req.Target,
		Username:      req.User,
		Password:      req.Password,
		Command:       req.Command,
		AllowHosts:    req.Policy.AllowHosts,
		AllowCommands: req.Policy.AllowCommands,
		Consent:       req.Consent,
		DryRun:        req.DryRun,
		Timeout:       req.Timeout,
	}

	res, err := remoteexec.Execute(ctx, internalReq)
	if err != nil {
		return contracts.RemoteExecResponse{}, err
	}

	return contracts.RemoteExecResponse{
		Output:  res.Output,
		Success: res.Success,
	}, nil
}

func (s *RemoteExecService) DryRun(ctx context.Context, req contracts.RemoteExecRequest) error {
	internalReq := remoteexec.Request{
		Transport:     req.Transport,
		Target:        req.Target,
		Username:      req.User,
		Password:      req.Password,
		Command:       req.Command,
		AllowHosts:    req.Policy.AllowHosts,
		AllowCommands: req.Policy.AllowCommands,
		Consent:       req.Consent,
		DryRun:        true,
		Timeout:       req.Timeout,
	}

	_, err := remoteexec.Execute(ctx, internalReq)
	return err
}
