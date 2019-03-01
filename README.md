```
CREATE TABLE patients (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL,
  phone TEXT NOT NULL
);

```

```
INSERT INTO patients (name, email, phone)
VALUES ('johne@test.com', 'John', '12345678');
```

# Graphql queries

#GET patients list
http://localhost:8000/patient?query={getPatients{id, name, email, phone}}


#GET a patient by ID
http://localhost:8000/patient?query={getPatient(id:1){id, name,email,phone}}


#CREATE a new patient
http://localhost:8000/patient?query=mutation+_{createPatient(name:"Andrew",
email: "andrew@test.com", 
phone: "890123490"){id,name,email,phone}}

#UPDATE an exisiting patient
http://localhost:8000/patient?query=mutation+_{updatePatient(id:1,phone: "3333333"){id,name,email,phone}}

#DELETE an exisiting patient
http://localhost:8000/patient?query=mutation+_{deletePatient(id:1){id,name,email,phone}}