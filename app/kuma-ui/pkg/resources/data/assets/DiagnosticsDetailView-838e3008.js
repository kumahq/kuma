import{_ as g}from"./CodeBlock.vue_vue_type_style_index_0_lang-dc96b835.js";import{E as f}from"./ErrorBlock-07b366f0.js";import{_ as h}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-6b4abfc9.js";import{d as x,a as t,o as i,b as s,w as a,e as n,m as C,f as b}from"./index-f5266944.js";import"./uniqueId-90cc9b93.js";import"./index-fce48c05.js";import"./TextWithCopyButton-ad2c2056.js";import"./CopyButton-3af07971.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-02a543f6.js";const T=x({__name:"DiagnosticsDetailView",setup(R){return(k,y)=>{const l=t("RouteTitle"),m=t("KCard"),p=t("AppView"),u=t("DataSource"),_=t("RouteView");return i(),s(_,{name:"diagnostics",params:{codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:c})=>[n(u,{src:"/config"},{default:a(({data:r,error:d})=>[n(p,{breadcrumbs:[{to:{name:"diagnostics"},text:c("diagnostics.routes.item.breadcrumbs")}]},{title:a(()=>[C("h1",null,[n(l,{title:c("diagnostics.routes.item.title")},null,8,["title"])])]),default:a(()=>[b(),n(m,null,{default:a(()=>[d?(i(),s(f,{key:0,error:d},null,8,["error"])):r===void 0?(i(),s(h,{key:1})):(i(),s(g,{key:2,id:"code-block-diagnostics","data-testid":"code-block-diagnostics",language:"json",code:JSON.stringify(r,null,2),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},null,8,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"]))]),_:2},1024)]),_:2},1032,["breadcrumbs"])]),_:2},1024)]),_:1})}}});export{T as default};
