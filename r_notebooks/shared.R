library(DBI)
library(tidyverse)

con <- DBI::dbConnect(odbc::odbc(),
                      Driver   = "PostgreSQL Unicode",
                      Server   = "localhost",
                      Database = "cheesecake",
                      UID      = "postgres",
                      PWD      = "postgres",
                      Port     = 5432)