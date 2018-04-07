# Back end service developed for my master project at The University of Bergen

## Endpoints

### OPTIONS '/session'
Post a new Session

### GET '/raw/sessions'
Get all unprocessed sessions

### GET '/raw/session/:id'
Get the unprocessed session with the given id

### POST '/raw/sessions'
Expects parameter "Finished" which is either 'false' or '1'
Get all unprocessed session with the given Finished value

### PUT '/raw/session/:id'
Update the session with the given id

### POST '/beacon'
Create a new Beacon

### GET '/beacons'
Get all beacons

### POST '/sessionbeacon'
Create a new SessionBeacon

### POST '/map'
Upload a .png image file

### GET '/debug/drop'
Drop all tables

### GET '/debug/drop/sessions'
Drop Session and SessionBeacon tables
