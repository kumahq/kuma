import{_ as E}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-jn80D5ag.js";import{d as x,r as n,o as r,m as s,w as a,b as t,e as w,P as R,ac as k,p as z}from"./index-DOjcqG3h.js";import"./CodeBlock-C-zU_pQm.js";import"./toYaml-DB9FPXFY.js";const b=x({__name:"ZoneEgressConfigView",setup(V){return(v,y)=>{const m=n("RouteTitle"),i=n("DataSource"),g=n("KCard"),_=n("AppView"),u=n("RouteView");return r(),s(u,{name:"zone-egress-config-view",params:{zoneEgress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:f})=>[t(m,{render:!1,title:f("zone-egresses.routes.item.navigation.zone-egress-config-view")},null,8,["title"]),w(),t(_,null,{default:a(()=>[t(g,null,{default:a(()=>[t(i,{src:`/zone-egresses/${e.params.zoneEgress}`},{default:a(({data:l,error:p})=>[p!==void 0?(r(),s(R,{key:0,error:p},null,8,["error"])):l===void 0?(r(),s(k,{key:1})):(r(),s(E,{key:2,resource:l.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{default:a(({copy:o,copying:C})=>[C?(r(),s(i,{key:0,src:`/zone-egresses/${e.params.zoneEgress}/as/kubernetes?no-store`,onChange:c=>{o(d=>d(c))},onError:c=>{o((d,h)=>h(c))}},null,8,["src","onChange","onError"])):z("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"]))]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{b as default};
