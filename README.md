# Movie Rater Rest Api

### Implementation
- GoFiber
- MySQL
- JWT

### MySQL Setup
```mysql
# Create movie_rater Database
CREATE DATABASE movie_rater;
USE movie_rater;

# movies Table
CREATE TABLE movies(
	id INTEGER UNSIGNED AUTO_INCREMENT,
   	title VARCHAR(75) NOT NULL,
    avgRating DECIMAL(2,1) NOT NULL DEFAULT 0.0 CHECK(avgRating BETWEEN 0.0 AND 5.0),
    CONSTRAINT id_pk PRIMARY KEY(id)
);

# users Table
CREATE TABLE users(
    id INTEGER UNSIGNED AUTO_INCREMENT,
    username VARCHAR(15) NOT NULL,
    email VARCHAR(35) NOT NULL,
    CONSTRAINT id_pk PRIMARY KEY(id)
);

# reviews Table
CREATE TABLE reviews(
	id INTEGER UNSIGNED AUTO_INCREMENT,
    rating DECIMAL(2,1) NOT NULL DEFAULT 0.0 CHECK(rating BETWEEN 0.0 AND 5.0)
    comment VARCHAR(500)
    movieId INTEGER UNSIGNED NOT NULL,
    userId INTEGER UNSIGNED NOT NULL,
    CONSTRAINT id_pk PRIMARY KEY(id),
    CONSTRAINT movieId_fk FOREIGN KEY(movieId) REFERENCES movies(id)
    	ON DELETE CASCADE
    	ON UPDATE RESTRICT,
    CONSTRAINT userId_fk FOREIGN KEY(userId) REFERENCES users(id)
    	ON DELETE CASCADE
    	ON UPDATE RESTRICT
);
```

