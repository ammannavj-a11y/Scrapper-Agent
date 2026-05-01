# UIVI-SaaS

go run gen.go
EMAIL_HASH: 8189b3050c07076141006f3ef99cf5ee1d5e5e9dd9ce1e329f77ec0e8fc90278
PASSWORD_HASH: $2a$10$5f5ruIpLo5./G7N6DshLjOw4nGQLfrbX6t/LQo59N21qhYJwxtjGK


INSERT INTO tenant_users (tenant_id, email, email_hash, password_hash, full_name)
VALUES (
  (SELECT id FROM tenants WHERE slug='my-tenant'),
  'admin@guvid.com',
  '8189b3050c07076141006f3ef99cf5ee1d5e5e9dd9ce1e329f77ec0e8fc90278',
  '$2a$10$5f5ruIpLo5./G7N6DshLjOw4nGQLfrbX6t/LQo59N21qhYJwxtjGK',
  'Admin User'
);
///


fix for db error

docker exec -i uivi-saas-db-1 psql -U uivi -d uivi_master -f /manual-seed.sql
docker compose down -v
docker compose up --build