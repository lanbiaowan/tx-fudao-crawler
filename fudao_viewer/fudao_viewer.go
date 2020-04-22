package main

func main() {


	      {{ range .List }}
        {{ $title := .Title }}
        {{/* .Title 上下文变量调用  func param1 param2 方法/函数调用  $.根节点变量调用 */}}
        <li>{{ $title }} -- {{ .CreatedAt.Format "2006-01-02 15:04:05" }} -- Author {{ $.Author }}</li>
      {{end}}
      
}
