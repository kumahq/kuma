import{E}from"./ErrorBlock-567378ca.js";import{_ as x}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-52bdda9b.js";import{_ as w}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-60c9f4a9.js";import{d as R,a as s,o as n,b as a,w as r,e as t,m as V,f as k,p as z}from"./index-5d5446a4.js";import"./index-fce48c05.js";import"./TextWithCopyButton-1669005d.js";import"./CopyButton-b62a1694.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-e798630e.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-be5bb8ae.js";import"./uniqueId-90cc9b93.js";import"./toYaml-4e00099e.js";const Q=R({__name:"ZoneEgressConfigView",setup(v){return(y,F)=>{const d=s("RouteTitle"),c=s("DataSource"),_=s("KCard"),g=s("AppView"),u=s("RouteView");return n(),a(u,{name:"zone-egress-config-view",params:{zoneEgress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:r(({route:e,t:f})=>[t(g,null,{title:r(()=>[V("h2",null,[t(d,{title:f("zone-egresses.routes.item.navigation.zone-egress-config-view")},null,8,["title"])])]),default:r(()=>[k(),t(_,null,{default:r(()=>[t(c,{src:`/zone-egresses/${e.params.zoneEgress}`},{default:r(({data:p,error:m})=>[m!==void 0?(n(),a(E,{key:0,error:m},null,8,["error"])):p===void 0?(n(),a(x,{key:1})):(n(),a(w,{key:2,resource:p.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{default:r(({copy:o,copying:h})=>[h?(n(),a(c,{key:0,src:`/zone-egresses/${e.params.zoneEgress}/as/kubernetes?no-store`,onChange:i=>{o(l=>l(i))},onError:i=>{o((l,C)=>C(i))}},null,8,["src","onChange","onError"])):z("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"]))]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{Q as default};
