package utils

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "mariadb-galera-cluster-operator.domain/api/v1"
	"strconv"
)

//func main()  {
//	cluster := &v1.MariaDBCluster{}
//	fmt.Println(getCommandLine(cluster))
//}

func MutateStatefulSet(cluster *v1.MariaDBCluster, sts *appsv1.StatefulSet) {
	sts.Labels = map[string]string{
		"app": "mariadb",
	}
	sts.Spec = appsv1.StatefulSetSpec{
		Replicas:    cluster.Spec.Size,
		ServiceName: cluster.Name,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "mariadb",
			}},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "mariadb",
				},
			},
			Spec: corev1.PodSpec{
				InitContainers: initContainers(cluster),
				Containers:     containers(cluster),
				Volumes: []corev1.Volume{
					corev1.Volume{
						Name: "conf",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
					corev1.Volume{
						Name: "start-up",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				},
			},
		},
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "data",
					Annotations: map[string]string{
						"volume.beta.kubernetes.io/storage-class": "nginx-nfs-storage",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("1Gi"),
						},
					},
				},
			},
		},
	}
}

func MutateHeadlessSVC(cluster *v1.MariaDBCluster, svc *corev1.Service) {
	svc.Labels = map[string]string{
		"app": "mariadb",
	}
	svc.Spec = corev1.ServiceSpec{
		ClusterIP: corev1.ClusterIPNone,
		Selector: map[string]string{
			"app": "mariadb",
		},
		Ports: []corev1.ServicePort{
			corev1.ServicePort{
				Name: "client",
				Port: 3306,
			},
		},
	}
}

func initContainers(cluster *v1.MariaDBCluster) []corev1.Container {
	var configCommandLine = getCommandLine(cluster)
	fmt.Println(configCommandLine)
	return []corev1.Container{
		corev1.Container{
			Name:  "mariadb-init",
			Image: "mariadb:10.3.10",
			VolumeMounts: []corev1.VolumeMount{
				corev1.VolumeMount{
					Name:      "conf",
					MountPath: "/mnt/conf.d",
				},
				corev1.VolumeMount{
					Name:      "start-up",
					MountPath: "/mnt/start-fzx",
				},
			},
			Command: []string{
				"/bin/bash",
				"-c",
				configCommandLine,
			},
		},
	}
}

func containers(cluster *v1.MariaDBCluster) []corev1.Container {
	return []corev1.Container{
		corev1.Container{
			Name:  "mariadb",
			Image: "mariadb:10.3.10",
			Ports: []corev1.ContainerPort{
				corev1.ContainerPort{
					Name:          "mariadb",
					ContainerPort: 3306,
				},
				corev1.ContainerPort{
					Name:          "communication",
					ContainerPort: 4567,
				},
				corev1.ContainerPort{
					Name:          "ist",
					ContainerPort: 4568,
				},
				corev1.ContainerPort{
					Name:          "sst",
					ContainerPort: 4444,
				},
			},
			Env: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "MYSQL_ALLOW_EMPTY_PASSWORD",
					Value: "1",
				},
				corev1.EnvVar{
					Name:  "TIMEZONE",
					Value: "Asia/Shanghai",
				},
				corev1.EnvVar{
					Name:  "MYSQL_INITDB_SKIP_TZINFO",
					Value: "yes",
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				corev1.VolumeMount{
					Name:      "data",
					MountPath: "/var/lib/mysql",
					SubPath:   "mysql",
				},
				corev1.VolumeMount{
					Name:      "conf",
					MountPath: "/etc/mysql/conf.d",
				},
				corev1.VolumeMount{
					Name:      "start-up",
					MountPath: "/etc/mysql/start-fzx",
				},
			},
			Command: []string{
				"/bin/bash",
				"-c",
				"/etc/mysql/start-fzx/start-up.sh",
			},
		},
	}
}

func getCommandLine(cluster *v1.MariaDBCluster) string {
	var commandLineComplete = "set -ex\n"
	commandLineComplete += "[[ `hostname` =~ -([0-9]+)$ ]] || exit 1\nordinal=${BASH_REMATCH[1]}\n"
	commandLineComplete += getMariaDBClusterString(cluster)
	commandLineComplete += "if [[ $ordinal -eq 0 ]]; then\n"
	commandLineComplete += "  echo \"if [[ -f /var/lib/mysql/grastate.dat ]]; then\" > /mnt/start-fzx/start-up.sh\n"
	commandLineComplete += "  echo \"  docker-entrypoint.sh mysqld\" >> /mnt/start-fzx/start-up.sh\n"
	commandLineComplete += "  echo \"else\" >> /mnt/start-fzx/start-up.sh\n"
	commandLineComplete += "  echo \"  docker-entrypoint.sh mysqld --wsrep-new-cluster\" >> /mnt/start-fzx/start-up.sh\n"
	commandLineComplete += "  echo \"fi\" >> /mnt/start-fzx/start-up.sh\n"
	commandLineComplete += "else\n"
	commandLineComplete += "  echo \"docker-entrypoint.sh mysqld\" > /mnt/start-fzx/start-up.sh\n"
	commandLineComplete += "fi\n"
	commandLineComplete += "chmod 777 /mnt/start-fzx/start-up.sh\n"
	return commandLineComplete
}

func getMariaDBClusterString(cluster *v1.MariaDBCluster) string {
	var clusterAddress = "wsrep_cluster_address=gcomm://"
	size := *cluster.Spec.Size
	for i := 1; i < int(size); i++ {
		clusterAddress += "mariadbcluster-sample-" + strconv.Itoa(i) + ".mariadbcluster-sample.default.svc.cluster.local,"
	}
	clusterAddress += "mariadbcluster-sample-0.mariadbcluster-sample.default.svc.cluster.local"
	var clusterCompleteString = "echo [galera] > /mnt/conf.d/MariaDBCluster.cnf\n"
	clusterCompleteString += "echo wsrep_on=ON >> /mnt/conf.d/MariaDBCluster.cnf\n"
	clusterCompleteString += "echo wsrep_provider=/usr/lib/galera/libgalera_smm.so >> /mnt/conf.d/MariaDBCluster.cnf\n"
	clusterAddress = "echo " + clusterAddress + " >> /mnt/conf.d/MariaDBCluster.cnf\n"
	clusterCompleteString += clusterAddress
	clusterCompleteString += "echo wsrep_cluster_name=mariadb >> /mnt/conf.d/MariaDBCluster.cnf\n"
	clusterCompleteString += "echo wsrep_node_name=mariadb-${ordinal} >> /mnt/conf.d/MariaDBCluster.cnf\n"
	clusterCompleteString += "echo binlog_format=ROW >> /mnt/conf.d/MariaDBCluster.cnf\n"
	clusterCompleteString += "echo default-storage-engine=innodb >> /mnt/conf.d/MariaDBCluster.cnf\n"
	clusterCompleteString += "echo innodb_autoinc_lock_mode=2 >> /mnt/conf.d/MariaDBCluster.cnf\n"
	clusterCompleteString += "echo bind-address=mariadbcluster-sample-${ordinal}.mariadbcluster-sample.default.svc.cluster.local >> /mnt/conf.d/MariaDBCluster.cnf\n"
	clusterCompleteString += "echo wsrep_slave_threads = 8 >> /mnt/conf.d/MariaDBCluster.cnf\n"
	clusterCompleteString += "echo innodb_flush_log_at_trx_commit = 0 >> /mnt/conf.d/MariaDBCluster.cnf\n"

	return clusterCompleteString
}
