import{C as g}from"./CodeBlock-1370eb60.js";import{E as f}from"./ErrorBlock-43034db1.js";import{_ as h}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-ea2db3d2.js";import{d as C,a as t,o as i,b as s,w as a,e as n,m as x,f as b}from"./index-a04e4171.js";import"./uniqueId-90cc9b93.js";import"./index-fce48c05.js";import"./TextWithCopyButton-7e4c909e.js";import"./CopyButton-882edec4.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-9aa5665b.js";const T=C({__name:"DiagnosticsDetailView",setup(k){return(R,y)=>{const l=t("RouteTitle"),m=t("KCard"),p=t("AppView"),u=t("DataSource"),_=t("RouteView");return i(),s(_,{name:"diagnostics",params:{codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:c})=>[n(u,{src:"/config"},{default:a(({data:r,error:d})=>[n(p,{breadcrumbs:[{to:{name:"diagnostics"},text:c("diagnostics.routes.item.breadcrumbs")}]},{title:a(()=>[x("h1",null,[n(l,{title:c("diagnostics.routes.item.title")},null,8,["title"])])]),default:a(()=>[b(),n(m,null,{default:a(()=>[d?(i(),s(f,{key:0,error:d},null,8,["error"])):r===void 0?(i(),s(h,{key:1})):(i(),s(g,{key:2,id:"code-block-diagnostics","data-testid":"code-block-diagnostics",language:"json",code:JSON.stringify(r,null,2),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},null,8,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"]))]),_:2},1024)]),_:2},1032,["breadcrumbs"])]),_:2},1024)]),_:1})}}});export{T as default};
