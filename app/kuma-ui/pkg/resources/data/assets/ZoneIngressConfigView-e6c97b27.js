import{E as x}from"./ErrorBlock-6c88077a.js";import{_ as w}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-bf912351.js";import{_ as E}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-8d89b4d0.js";import{d as R,a as r,o as s,b as a,w as n,e as t,m as V,f as k,p as z}from"./index-f385383f.js";import"./index-fce48c05.js";import"./TextWithCopyButton-66e03f53.js";import"./CopyButton-bd8ef627.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-f04c574e.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-24482f8c.js";import"./uniqueId-90cc9b93.js";import"./toYaml-4e00099e.js";const K=R({__name:"ZoneIngressConfigView",setup(v){return(y,F)=>{const d=r("RouteTitle"),c=r("DataSource"),_=r("KCard"),g=r("AppView"),u=r("RouteView");return s(),a(u,{name:"zone-ingress-config-view",params:{zoneIngress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:n(({route:e,t:f})=>[t(g,null,{title:n(()=>[V("h2",null,[t(d,{title:f("zone-ingresses.routes.item.navigation.zone-ingress-config-view")},null,8,["title"])])]),default:n(()=>[k(),t(_,null,{default:n(()=>[t(c,{src:`/zone-ingresses/${e.params.zoneIngress}`},{default:n(({data:p,error:m})=>[m!==void 0?(s(),a(x,{key:0,error:m},null,8,["error"])):p===void 0?(s(),a(w,{key:1})):(s(),a(E,{key:2,resource:p.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{default:n(({copy:o,copying:h})=>[h?(s(),a(c,{key:0,src:`/zone-ingresses/${e.params.zoneIngress}/as/kubernetes?no-store`,onChange:i=>{o(l=>l(i))},onError:i=>{o((l,C)=>C(i))}},null,8,["src","onChange","onError"])):z("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"]))]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{K as default};
