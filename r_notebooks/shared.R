library(DBI)
library(tidyverse)

#con <- DBI::dbConnect(odbc::odbc(),
#                      Driver   = "PostgreSQL Unicode",
#                      Server   = "localhost",
#                      Database = "cheesecake",
#                      UID      = "cheese",
#                      PWD      = "cheesepass4279",
#                      Port     = 5432)

con <- dbConnect(RPostgres::Postgres(),
                 dbname = 'cheesecake', 
                 host = 'localhost', # i.e. 'ec2-54-83-201-96.compute-1.amazonaws.com'
                 port = 5432, # or any other port specified by your DBA
                 user = 'cheese',
                 password = 'cheesepass4279')