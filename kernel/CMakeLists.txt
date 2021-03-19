cmake_minimum_required(VERSION 3.0.0 FATAL_ERROR)

project(kernel VERSION 0.1.0 LANGUAGES C)
set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)

set(HEADERS
        ml_crash_detector.h
        state.h
        communication.h)
set(SOURCE
        ml_crash_detector.c
        state.c
        communication.c)

# Find the kernel release
execute_process(
        COMMAND uname -r
        OUTPUT_VARIABLE KERNEL_RELEASE
        OUTPUT_STRIP_TRAILING_WHITESPACE
)

# Find the headers
find_path(KERNELHEADERS_DIR
        Module.symvers
        PATHS /lib/modules/${KERNEL_RELEASE}/build
        )

message(STATUS "Kernel release: ${KERNEL_RELEASE}")
message(STATUS "Kernel headers: ${KERNELHEADERS_DIR}")
set(KERNELHEADERS_INCLUDE_DIRS
        ${KERNELHEADERS_DIR}/include
        ${KERNELHEADERS_DIR}/arch/x86/include
        CACHE PATH "Kernel headers include dirs"
        )
include_directories(${CUSTOM_INCLUDES})
include_directories(${KERNELHEADERS_INCLUDE_DIRS})

add_executable(kernel ${HEADERS} ${SOURCE})