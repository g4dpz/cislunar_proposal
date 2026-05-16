cmake_minimum_required(VERSION 3.16)
project(kiss_cla_plugin CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

add_library(kiss_cla_plugin SHARED
    kiss_cla_plugin.cpp
)

target_include_directories(kiss_cla_plugin PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}
    ${HDTN_INCLUDE_DIR}
)

install(TARGETS kiss_cla_plugin
    LIBRARY DESTINATION lib
)
