package cmd

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/domain"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
)

func (a *App) newServiceCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "service", Short: "Naming: list/instances/register/deregister/up/down", RunE: helpOnly}
	cmd.AddCommand(
		a.newServiceListCmd(),
		a.newServiceInstancesCmd(),
		a.newServiceRegisterCmd(),
		a.newServiceDeregisterCmd(),
		a.newServiceUpCmd(),
		a.newServiceDownCmd(),
	)
	return cmd
}

func (a *App) newServiceListCmd() *cobra.Command {
	var ns string
	var page, pageSize int
	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.ServiceUseCase{Port: a.client}
			out, err := uc.List(ns, page, pageSize)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			return a.ok(cmd, out)
		},
	}
	cmd.Flags().StringVar(&ns, "namespace", "", "namespace id")
	cmd.Flags().IntVar(&page, "page", 1, "pageNo")
	cmd.Flags().IntVar(&pageSize, "page-size", 100, "pageSize")
	mark(cmd, risk.Read, "List services")
	return cmd
}

func (a *App) newServiceInstancesCmd() *cobra.Command {
	var service, ns, group string
	cmd := &cobra.Command{
		Use: "instances",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.ServiceUseCase{Port: a.client}
			out, err := uc.Instances(service, ns, group)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			return a.ok(cmd, out)
		},
	}
	cmd.Flags().StringVar(&service, "service", "", "service name (required)")
	cmd.Flags().StringVar(&ns, "namespace", "", "namespace id")
	cmd.Flags().StringVar(&group, "group", "DEFAULT_GROUP", "groupName")
	_ = cmd.MarkFlagRequired("service")
	mark(cmd, risk.Read, "List instances of a service")
	return cmd
}

func (a *App) addInstanceFlags(cmd *cobra.Command, service, ip, ns, group, cluster *string, port *int, ephemeral *bool, persistent *bool) {
	cmd.Flags().StringVar(service, "service", "", "service name (required)")
	cmd.Flags().StringVar(ip, "ip", "", "instance ip (required)")
	cmd.Flags().IntVar(port, "port", 0, "instance port (required)")
	cmd.Flags().StringVar(ns, "namespace", "", "namespace id")
	cmd.Flags().StringVar(group, "group", "DEFAULT_GROUP", "groupName")
	cmd.Flags().StringVar(cluster, "cluster", "DEFAULT", "clusterName")
	cmd.Flags().BoolVar(ephemeral, "ephemeral", true, "ephemeral instance (default true)")
	cmd.Flags().BoolVar(persistent, "persistent", false, "force persistent instance (overrides --ephemeral)")
	_ = cmd.MarkFlagRequired("service")
	_ = cmd.MarkFlagRequired("ip")
	_ = cmd.MarkFlagRequired("port")
}

func resolveEphemeral(ephemeral, persistent bool, cmd *cobra.Command) bool {
	if persistent {
		return false
	}
	if cmd.Flags().Changed("ephemeral") {
		return ephemeral
	}
	return true // default ephemeral
}

func (a *App) newServiceRegisterCmd() *cobra.Command {
	var service, ip, ns, group, cluster string
	var port int
	var ephemeral, persistent bool
	cmd := &cobra.Command{
		Use:  "register",
		Long: "Register an instance. Default ephemeral=true; use --persistent for durable instances.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			eph := resolveEphemeral(ephemeral, persistent, cmd)
			would := map[string]any{
				"action": "service.register", "service": service, "ip": ip, "port": port,
				"namespace": ns, "group": group, "cluster": cluster, "ephemeral": eph,
			}
			stop, err := a.gateHigh(cmd, "service.register", would)
			if stop || err != nil {
				return err
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.ServiceUseCase{Port: a.client}
			out, err := uc.Register(service, ip, port, ns, group, cluster, eph)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, out)
		},
	}
	a.addInstanceFlags(cmd, &service, &ip, &ns, &group, &cluster, &port, &ephemeral, &persistent)
	mark(cmd, risk.High, "Register instance (ephemeral by default)")
	return cmd
}

func (a *App) newServiceDeregisterCmd() *cobra.Command {
	var service, ip, ns, group, cluster string
	var port int
	var ephemeral, persistent bool
	cmd := &cobra.Command{
		Use: "deregister",
		RunE: func(cmd *cobra.Command, _ []string) error {
			eph := resolveEphemeral(ephemeral, persistent, cmd)
			would := map[string]any{
				"action": "service.deregister", "service": service, "ip": ip, "port": port,
				"namespace": ns, "group": group, "cluster": cluster, "ephemeral": eph,
			}
			stop, err := a.gateHigh(cmd, "service.deregister", would)
			if stop || err != nil {
				return err
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.ServiceUseCase{Port: a.client}
			out, err := uc.Deregister(service, ip, port, ns, group, cluster, eph)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, out)
		},
	}
	a.addInstanceFlags(cmd, &service, &ip, &ns, &group, &cluster, &port, &ephemeral, &persistent)
	mark(cmd, risk.High, "Deregister instance")
	return cmd
}

func (a *App) newServiceUpCmd() *cobra.Command {
	var service, ip, ns, group, cluster string
	var port int
	var ephemeral, persistent bool
	cmd := &cobra.Command{
		Use: "up",
		RunE: func(cmd *cobra.Command, _ []string) error {
			eph := resolveEphemeral(ephemeral, persistent, cmd)
			would := map[string]any{
				"action": "service.up", "service": service, "ip": ip, "port": port,
				"namespace": ns, "group": group, "cluster": cluster, "ephemeral": eph,
			}
			stop, err := a.gateHigh(cmd, "service.up", would)
			if stop || err != nil {
				return err
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.ServiceUseCase{Port: a.client}
			out, err := uc.Up(service, ip, port, ns, group, cluster, eph)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, out)
		},
	}
	a.addInstanceFlags(cmd, &service, &ip, &ns, &group, &cluster, &port, &ephemeral, &persistent)
	mark(cmd, risk.High, "Enable instance (online)")
	return cmd
}

func (a *App) newServiceDownCmd() *cobra.Command {
	var service, ip, ns, group, cluster string
	var port int
	var ephemeral, persistent bool
	cmd := &cobra.Command{
		Use: "down",
		RunE: func(cmd *cobra.Command, _ []string) error {
			eph := resolveEphemeral(ephemeral, persistent, cmd)
			would := map[string]any{
				"action": "service.down", "service": service, "ip": ip, "port": port,
				"namespace": ns, "group": group, "cluster": cluster, "ephemeral": eph,
			}
			stop, err := a.gateHigh(cmd, "service.down", would)
			if stop || err != nil {
				return err
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.ServiceUseCase{Port: a.client}
			out, err := uc.Down(service, ip, port, ns, group, cluster, eph)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, out)
		},
	}
	a.addInstanceFlags(cmd, &service, &ip, &ns, &group, &cluster, &port, &ephemeral, &persistent)
	mark(cmd, risk.High, "Disable instance (offline)")
	return cmd
}

var _ = strconv.FormatBool
