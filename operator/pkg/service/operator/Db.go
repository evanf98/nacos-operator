package operator

import (
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	nacosgroupv1alpha1 "nacos.io/nacos-operator/api/v1alpha1"
)

func setDB(nacos *nacosgroupv1alpha1.Nacos) {
	// 默认设置内置数据库
	if nacos.Spec.Database.TypeDatabase == "" {
		nacos.Spec.Database.TypeDatabase = "embedded"
	}

	switch nacos.Spec.Database.TypeDatabase {
	case "mysql":
		setMysqlDB(nacos)
	case "postgresql":
		setPgsqlDB(nacos)
	default:
		nacos.Spec.Database.TypeDatabase = "embedded"
	}

}

func setMysqlDB(nacos *nacosgroupv1alpha1.Nacos) {
	// mysql设置默认值
	if nacos.Spec.Database.DBHost == "" {
		nacos.Spec.Database.DBHost = "127.0.0.1"
	}
	if nacos.Spec.Database.DBUser == "" {
		nacos.Spec.Database.DBUser = "root"
	}
	if nacos.Spec.Database.DBName == "" {
		nacos.Spec.Database.DBName = "nacos"
	}
	if nacos.Spec.Database.DBPassword == "" {
		nacos.Spec.Database.DBPassword = "123456"
	}
	if nacos.Spec.Database.DBPort == "" {
		nacos.Spec.Database.DBPort = "3306"
	}
}
func setPgsqlDB(nacos *nacosgroupv1alpha1.Nacos) {
	if nacos.Spec.Database.DBHost == "" {
		nacos.Spec.Database.DBHost = "127.0.0.1"
	}
	if nacos.Spec.Database.DBUser == "" {
		nacos.Spec.Database.DBUser = "postgres"
	}
	if nacos.Spec.Database.DBName == "" {
		nacos.Spec.Database.DBName = "nacos"
	}
	if nacos.Spec.Database.DBPassword == "" {
		nacos.Spec.Database.DBPassword = "123456"
	}
	if nacos.Spec.Database.DBPort == "" {
		nacos.Spec.Database.DBPort = "5432"
	}
}

func CreateMySQLDbJob(nacos *nacosgroupv1alpha1.Nacos) (job *batchv1.Job) {
	job = &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nacos.Name + "-mysql-sql-init",
			Namespace: nacos.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: nacos.Namespace,
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name:  "mysql-ping-database",
							Image: nacos.Spec.DBInitImage,
							Env: []v1.EnvVar{
								{
									Name:  "MYSQL_HOST",
									Value: nacos.Spec.Database.DBHost,
								},
								{
									Name:  "MYSQL_DB",
									Value: nacos.Spec.Database.DBName,
								},
								{
									Name:  "MYSQL_PORT",
									Value: nacos.Spec.Database.DBPort,
								},
								{
									Name:  "MYSQL_USER",
									Value: nacos.Spec.Database.DBUser,
								},
								{
									Name:  "MYSQL_PASS",
									Value: nacos.Spec.Database.DBPassword,
								},
							},
							Command: []string{
								"/bin/sh",
								"-c",
								"while ! mysqladmin ping --host=\"${MYSQL_HOST}\" --port=\"${MYSQL_PORT}\" --user=\"${MYSQL_USER}\" --password=\"${MYSQL_PASS}\" ; do echo \"check mysql\"; sleep 1; done",
							},
						},
						{
							Name:  "mysql-create-database",
							Image: nacos.Spec.DBInitImage,
							Env: []v1.EnvVar{
								{
									Name:  "MYSQL_HOST",
									Value: nacos.Spec.Database.DBHost,
								},
								{
									Name:  "MYSQL_DB",
									Value: nacos.Spec.Database.DBName,
								},
								{
									Name:  "MYSQL_PORT",
									Value: nacos.Spec.Database.DBPort,
								},
								{
									Name:  "MYSQL_USER",
									Value: nacos.Spec.Database.DBUser,
								},
								{
									Name:  "MYSQL_PASS",
									Value: nacos.Spec.Database.DBPassword,
								},
							},
							Command: []string{
								"/bin/sh",
								"-c",
								"until mysql -u\"${MYSQL_USER}\" -p\"${MYSQL_PASS}\" -h\"${MYSQL_HOST}\" -P\"${MYSQL_PORT}\" -e\"create database if not exists \"${MYSQL_DB}\"\"; do echo waiting for database creation...; sleep 2; done;",
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:  "mysql-sql-init",
							Image: nacos.Spec.DBInitImage,
							Env: []v1.EnvVar{
								{
									Name:  "MYSQL_HOST",
									Value: nacos.Spec.Database.DBHost,
								},
								{
									Name:  "MYSQL_DB",
									Value: nacos.Spec.Database.DBName,
								},
								{
									Name:  "MYSQL_PORT",
									Value: nacos.Spec.Database.DBPort,
								},
								{
									Name:  "MYSQL_USER",
									Value: nacos.Spec.Database.DBUser,
								},
								{
									Name:  "MYSQL_PASS",
									Value: nacos.Spec.Database.DBPassword,
								},
								{
									Name: "SQL_SCRIPT",
									ValueFrom: &v1.EnvVarSource{
										ConfigMapKeyRef: &v1.ConfigMapKeySelector{
											LocalObjectReference: v1.LocalObjectReference{
												Name: nacos.Name + "-mysql-sql-init",
											},
											Key: "SQL_SCRIPT",
										},
									},
								},
							},
							Command: []string{
								"/bin/sh",
								"-c",
								"mysql -u\"${MYSQL_USER}\" -p\"${MYSQL_PASS}\" -h\"${MYSQL_HOST}\" -P\"${MYSQL_PORT}\" -D\"${MYSQL_DB}\" -e\"${SQL_SCRIPT}\";",
							},
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}
	return
}

func CreatePgSQLDbJob(nacos *nacosgroupv1alpha1.Nacos) (job *batchv1.Job) {
	job = &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nacos.Name + "-pgsql-sql-init",
			Namespace: nacos.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: nacos.Namespace,
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name:  "pgsql-check-database",
							Image: nacos.Spec.DBInitImage,
							Env: []v1.EnvVar{
								{
									Name:  "PGSQL_HOST",
									Value: nacos.Spec.Database.DBHost,
								},
								{
									Name:  "PGSQL_DB",
									Value: nacos.Spec.Database.DBName,
								},
								{
									Name:  "PGSQL_PORT",
									Value: nacos.Spec.Database.DBPort,
								},
								{
									Name:  "PGSQL_USER",
									Value: nacos.Spec.Database.DBUser,
								},
								{
									Name:  "PGPASSWORD",
									Value: nacos.Spec.Database.DBPassword,
								},
							},
							Command: []string{
								"/bin/sh",
								"-c",
								"while ! pg_isready -h \"${PGSQL_HOST}\" -p \"${PGSQL_PORT}\" -U \"${PGSQL_USER}\" -d postgres 2>&1 >> /dev/null; do echo \"check pgsql\"; sleep 1; done",
							},
						},
						{
							Name:  "pgsql-create-database",
							Image: nacos.Spec.DBInitImage,
							Env: []v1.EnvVar{
								{
									Name:  "PGSQL_HOST",
									Value: nacos.Spec.Database.DBHost,
								},
								{
									Name:  "PGSQL_DB",
									Value: nacos.Spec.Database.DBName,
								},
								{
									Name:  "PGSQL_PORT",
									Value: nacos.Spec.Database.DBPort,
								},
								{
									Name:  "PGSQL_USER",
									Value: nacos.Spec.Database.DBUser,
								},
								{
									Name:  "PGPASSWORD",
									Value: nacos.Spec.Database.DBPassword,
								},
							},
							Command: []string{
								"/bin/sh",
								"-c",
								"until psql -h \"${PGSQL_HOST}\" -p \"${PGSQL_PORT}\" -U \"${PGSQL_USER}\" -d postgres -c\"create database ${PGSQL_DB}\"; do echo waiting for database creation...; sleep 2; done;",
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:  "pgsql-sql-init",
							Image: nacos.Spec.DBInitImage,
							Env: []v1.EnvVar{
								{
									Name:  "PGSQL_HOST",
									Value: nacos.Spec.Database.DBHost,
								},
								{
									Name:  "PGSQL_DB",
									Value: nacos.Spec.Database.DBName,
								},
								{
									Name:  "PGSQL_PORT",
									Value: nacos.Spec.Database.DBPort,
								},
								{
									Name:  "PGSQL_USER",
									Value: nacos.Spec.Database.DBUser,
								},
								{
									Name:  "PGPASSWORD",
									Value: nacos.Spec.Database.DBPassword,
								},
								{
									Name: "SQL_SCRIPT",
									ValueFrom: &v1.EnvVarSource{
										ConfigMapKeyRef: &v1.ConfigMapKeySelector{
											LocalObjectReference: v1.LocalObjectReference{
												Name: nacos.Name + "-postgresql-sql-init",
											},
											Key: "SQL_SCRIPT",
										},
									},
								},
							},
							Command: []string{
								"/bin/sh",
								"-c",
								"echo \"${SQL_SCRIPT}\" > /tmp/pgsql.sql | psql -h \"${PGSQL_HOST}\" -p \"${PGSQL_PORT}\" -U \"${PGSQL_USER}\" -d ${PGSQL_DB} -f /tmp/pgsql.sql",
							},
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}
	return
}
