package cluster

import (
	"fmt"

	api "github.com/xoe-labs/odoo-operator/pkg/apis/odoo/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// Volume Names
	configVolName      = "config"
	persistenceVolName = "persistence"
	backupVolName      = "backup"

	// Ports and Port Names
	clientPortName      = "client-port"
	clientPort          = 8069
	longpollingPortName = "longpolling-port"
	longpollingPort     = 8072
)

func deploymentsForOdooTrack(cr *api.OdooCluster, trackSpec *api.TrackSpec) []*appsv1.Deployment {
	selector := selectorForOdooCluster(cr.GetName())
	volumes := []v1.Volume{
		{
			Name: configVolName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: configMapNameForOdoo(cr),
					},
				},
			},
		},
	}

	for s := range cr.Spec.PVCSpecs {
		vol := v1.Volume{
			// kubernetes.io/pvc-protection
			Name: volumeNameForOdoo(cr, s),
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: volumeNameForOdoo(cr, s),
					ReadOnly:  false,
				},
			},
		}
		append(volumes, vol)

	}

	securityContext := &v1.PodSecurityContext{
		RunAsUser:    func(i int64) *int64 { return &i }(9001),
		RunAsNonRoot: func(b bool) *bool { return &b }(true),
		FSGroup:      func(i int64) *int64 { return &i }(9001),
	}

	var deployments []*appsv1.Deployment

	for _, tierSpec := range cr.Tiers {
		objectMeta := metav1.ObjectMeta{
			Name:      getFullName(cr, trackSpec, tierSpec),
			Namespace: cr.GetNamespace(),
			Labels:    labelsWithTrackAndTier(selector, trackSpec, tierSpec),
		}

		podTempl := v1.PodTemplateSpec{
			ObjectMeta: meta,
			Spec: v1.PodSpec{
				Containers: []v1.Container{odooContainer(cr, trackSpec, tierSpec)},
				// Containers: []v1.Container{odooContainer(cr), odooMonitoringContainer(cr)},
				Volumes:         volumes,
				SecurityContext: securityContext,
			},
		}

		d := &appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			ObjectMeta: objectMeta,
			Spec: appsv1.DeploymentSpec{
				Replicas: &tierSpec.Replicas,
				Selector: &metav1.LabelSelector{MatchLabels: selector},
				Template: podTempl,
				Strategy: appsv1.DeploymentStrategy{
					Type: appsv1.RollingUpdateDeploymentStrategyType,
					RollingUpdate: &appsv1.RollingUpdateDeployment{
						MaxUnavailable: func(a intstr.IntOrString) *intstr.IntOrString { return &a }(intstr.FromInt(1)),
						MaxSurge:       func(a intstr.IntOrString) *intstr.IntOrString { return &a }(intstr.FromInt(1)),
					},
				},
			},
		}
		addOwnerRefToObject(d, asOwner(cr))
		append(deployments, d)

	}
	return deployments

}

func odooContainer(cr *api.OdooCluster, trackSpec *api.TrackSpec, tierSpec *api.TierSpec) v1.Container {

	command := getContainerCommand(tierSpec)
	ports := getContainerPorts(tierSpec)
	volumes := []v1.VolumeMount{
		{
			Name:      configVolName,
			MountPath: filepath.Dir(odooConfigDir),
		},
		{
			Name:      persistenceVolName,
			MountPath: filepath.Dir(odooPersistenceDir),
		},
		{
			Name:      backupVolName,
			MountPath: filepath.Dir(odooBackupDir),
		},
	}

	c := v1.Container{
		Name:         getFullName(cr, trackSpec, tierSpec),
		Image:        getImageName(trackSpec.Image),
		Command:      command,
		VolumeMounts: volumes,
		Ports:        ports,
		LivenessProbe: &v1.Probe{
			Handler: v1.Handler{
				Exec: &v1.ExecAction{
					Command: []string{
						"curl",
						"--connect-timeout", "5",
						"--max-time", "10",
						"-k", "-s",
						fmt.Sprintf("https://localhost:%d/web", clientPortName),
					},
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      10,
			PeriodSeconds:       60,
			FailureThreshold:    3,
		},
		ReadinessProbe: &v1.Probe{
			Handler: v1.Handler{
				HTTPGet: &v1.HTTPGetAction{
					Path:   "/web",
					Port:   intstr.FromInt(clientPortName),
					Scheme: v1.URISchemeHTTPS,
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      10,
			PeriodSeconds:       10,
			FailureThreshold:    3,
		},
	}
	switch n := tierSpec.Name.(type) {
	case *api.ServerTier:
		c.LivenessProbe = &v1.Probe{
			Handler: v1.Handler{
				Exec: &v1.ExecAction{
					Command: []string{
						"curl",
						"--connect-timeout", "5",
						"--max-time", "10",
						"-k", "-s",
						fmt.Sprintf("https://localhost:%d/web", clientPort),
					},
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      10,
			PeriodSeconds:       60,
			FailureThreshold:    3,
		}
		c.ReadinessProbe = &v1.Probe{
			Handler: v1.Handler{
				HTTPGet: &v1.HTTPGetAction{
					Path:   "/web",
					Port:   intstr.FromInt(clientPort),
					Scheme: v1.URISchemeHTTPS,
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      10,
			PeriodSeconds:       10,
			FailureThreshold:    3,
		}
	case *api.LongpollingTier:
		c.LivenessProbe = &v1.Probe{
			Handler: v1.Handler{
				Exec: &v1.ExecAction{
					Command: []string{
						"curl",
						"--connect-timeout", "5",
						"--max-time", "10",
						"-k", "-s",
						fmt.Sprintf("https://localhost:%d/web", longpollingPort),
					},
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      10,
			PeriodSeconds:       60,
			FailureThreshold:    3,
		}
		c.ReadinessProbe = &v1.Probe{
			Handler: v1.Handler{
				HTTPGet: &v1.HTTPGetAction{
					Path:   "/web",
					Port:   intstr.FromInt(longpollingPort),
					Scheme: v1.URISchemeHTTPS,
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      10,
			PeriodSeconds:       10,
			FailureThreshold:    3,
		}
	}
}

func getContainerCommand(tierSpec *api.TierSpec) []string {
	switch n := tierSpec.Name.(type) {
	case *api.ServerTier:
		return []string{"--config", odooConfigDir}
		// return []string{"--config", odooConfigDir, "--tier", api.ServerTier}
	case *api.CronTier:
		return []string{"--config", odooConfigDir}
		// return []string{"--config", odooConfigDir, "--tier", api.CronTier}
	case *api.BackgroundTier:
		return []string{"--config", odooConfigDir}
		// return []string{"--config", odooConfigDir, "--tier", api.BackgroundTier}
	case *api.LongpollingTier:
		return []string{"--config", odooConfigDir}
		// return []string{"--config", odooConfigDir, "--tier", api.LongpollingTier}
	}
}

func getContainerPorts(tierSpec *api.TierSpec) []v1.ContainerPort {
	switch n := tierSpec.Name.(type) {
	case *api.ServerTier:
		return []v1.ContainerPort{{
			Name:          clientPortName,
			ContainerPort: int32(clientPort),
		}}
	case *api.CronTier:
		return []v1.ContainerPort{}
	case *api.BackgroundTier:
		return []v1.ContainerPort{}
	case *api.LongpollingTier:
		return []v1.ContainerPort{{
			Name:          longpollingPortName,
			ContainerPort: int32(longpollingPort),
		}}
	}
}
