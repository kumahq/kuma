import{_ as f}from"./CodeBlock.vue_vue_type_style_index_0_lang-a2e2bc37.js";import{d as g,a as e,o as t,b as a,w as o,e as s,p as b,f as h,H as y,I as k}from"./index-646486ee.js";const B=g({__name:"DiagnosticsView",setup(V){return(w,C)=>{const d=e("RouteTitle"),l=e("KCard"),u=e("AppView"),_=e("DataSource"),m=e("RouteView");return t(),a(m,{name:"diagnostics",params:{codeSearch:""}},{default:o(({route:c,t:n})=>[s(_,{src:"/config"},{default:o(({data:i,error:r})=>[s(u,{breadcrumbs:[{to:{name:"diagnostics"},text:n("diagnostics.routes.item.breadcrumbs")}]},{title:o(()=>[b("h1",null,[s(d,{title:n("diagnostics.routes.item.title")},null,8,["title"])])]),default:o(()=>[h(),s(l,null,{body:o(()=>[r?(t(),a(y,{key:0,error:r},null,8,["error"])):i===void 0?(t(),a(k,{key:1})):(t(),a(f,{key:2,id:"code-block-diagnostics","data-testid":"code-block-diagnostics",language:"json",code:JSON.stringify(i,null,2),"is-searchable":"",query:c.params.codeSearch,onQueryChange:p=>c.update({codeSearch:p})},null,8,["code","query","onQueryChange"]))]),_:2},1024)]),_:2},1032,["breadcrumbs"])]),_:2},1024)]),_:1})}}});export{B as default};
