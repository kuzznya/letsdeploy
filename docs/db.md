# Database schema

```plantuml
@startuml
' hide the spot
hide circle

' avoid problems with angled crows feet
skinparam linetype ortho

entity project {
  id
  invite_code
}

entity project_participant {
  id 
  project_id: <<FK project(id)>>
  username
}

entity service {
  id
  project_id: <<FK project(id)>>
  name
  image
  port
}

entity managed_service {
  id
  project_id: <<FK project(id)>>
  name
  type: postgres|mysql|rabbitmq|redis
  auth_secret_id <<FK secret(id)>>
}

entity env_var {
  id
  service_id <<FK service(id)>>
  name
  value
  secret_id <<FK secret(id)>>
}

entity secret {
  id
  project_id <<FK project(id)>>
  name
  managed_service_id <<FK managed_service(id)>>
}

project }o.up.o{ project_participant
project ||..o{ service
project ||..o{ managed_service
service ||..o{ env_var
env_var |o.right.o| secret
project ||..o{ secret
secret ||.up.o{ managed_service

@enduml
```
