section .data
global n
	n dd 0
global m
	m dd 0
global k
	k dd 3
	.L1 db "abc",10
section .text
global print
print:
push ebp
mov ebp, esp
sub esp, 0
.L22:
mov esp, ebp
pop ebp
ret
global fun
fun:
push ebp
mov ebp, esp
sub esp, 4
mov eax, 10
push eax
call print
add esp, 4
mov [ebp-4], eax
mov eax, 10
jmp .L23
.L23:
mov esp, ebp
pop ebp
ret
global main
main:
push ebp
mov ebp, esp
sub esp, 48
mov eax, 0
mov [ebp-4], eax
mov eax, 0
mov [ebp-8], eax
mov eax, 0
mov [ebp-12], eax
mov eax, [ebp-8]
mov [ebp-16], eax
mov eax, [ebp-8]
mov ebx, 1
sub eax, ebx
mov [ebp-8], eax
mov eax, [ebp-16]
mov [ebp-12], eax
.L4:
mov eax, [ebp-4]
mov ebx, 0
mov ecx, 0
cmp eax, ebx
setl cl
mov [ebp-20], ecx
mov eax, [ebp-20]
cmp eax, 0
je .L5
jmp .L4
.L5:
mov eax, [ebp-4]
mov ebx, [ebp-8]
mov ecx, 0
cmp eax, ebx
setl cl
mov [ebp-24], ecx
mov eax, [ebp-24]
cmp eax, 0
je .L7
mov eax, [ebp-4]
mov ebx, [ebp-8]
mov ecx, 0
cmp eax, ebx
setge cl
mov [ebp-28], ecx
mov eax, [ebp-28]
mov [ebp-4], eax
.L7:
mov eax, [ebp-4]
mov bl, 97
cmp eax, ebx
jne .L12
mov eax, [ebp-8]
mov [ebp-4], eax
.L12:
jmp .L11
.L11:
mov eax, 100
mov [ebp-32], eax
mov eax, 0
mov [ebp-36], eax
.L13:
mov eax, [ebp-36]
mov ebx, [ebp-32]
mov ecx, 0
cmp eax, ebx
setl cl
mov [ebp-40], ecx
mov eax, [ebp-40]
cmp eax, 0
je .L16
jmp .L14
.L15:
mov eax, [ebp-36]
mov [ebp-44], eax
mov eax, [ebp-36]
mov ebx, 1
add eax, ebx
mov [ebp-36], eax
jmp .L13
.L14:
mov eax, [ebp-32]
mov ebx, 0
mov ecx, 0
cmp eax, ebx
sete cl
mov [ebp-48], ecx
mov eax, [ebp-48]
cmp eax, 0
je .L19
jmp .L15
.L19:
jmp .L15
.L16:
.L2:
mov esp, ebp
pop ebp
ret
