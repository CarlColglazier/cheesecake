library(DBI)
library(tidyverse)

con <- DBI::dbConnect(odbc::odbc(),
                      Driver   = "PostgreSQL Unicode",
                      Server   = "localhost",
                      Database = "cheesecake",
                      UID      = "cheese",
                      PWD      = "cheesepass4279",
                      Port     = 5432)