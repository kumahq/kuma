import{_ as C}from"./CodeBlock.vue_vue_type_style_index_0_lang-c74a2004.js";import{d as x,a,o as i,b as r,w as o,e as d,$ as w,p as u,f as c,t as R,c as g,F as z,J as V}from"./index-d015481a.js";const F=["data-testid","innerHTML"],v=x({__name:"ZoneConfigView",props:{data:{},notifications:{default:()=>[]}},setup(m){const s=m;return(M,b)=>{const f=a("RouteTitle"),_=a("KAlert"),h=a("KCard"),k=a("AppView"),y=a("RouteView");return i(),r(y,{name:"zone-cp-config-view",params:{zone:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:n,t:l})=>[d(k,null,w({title:o(()=>[u("h2",null,[d(f,{title:l("zone-cps.routes.item.navigation.zone-cp-config-view")},null,8,["title"])])]),default:o(()=>[c(),c(),d(h,null,{default:o(()=>{var e,p;return[Object.keys(((e=s.data.zoneInsight)==null?void 0:e.config)??{}).length>0?(i(),r(C,{key:0,id:"code-block-zone-config",language:"json",code:JSON.stringify(((p=s.data.zoneInsight)==null?void 0:p.config)??{},null,2),"is-searchable":"",query:n.params.codeSearch,"is-filter-mode":n.params.codeFilter==="true","is-reg-exp-mode":n.params.codeRegExp==="true",onQueryChange:t=>n.update({codeSearch:t}),onFilterModeChange:t=>n.update({codeFilter:t}),onRegExpModeChange:t=>n.update({codeRegExp:t})},null,8,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])):(i(),r(_,{key:1,class:"mt-4","data-testid":"warning-no-subscriptions",appearance:"warning"},{alertMessage:o(()=>[c(R(l("zone-cps.detail.no_subscriptions")),1)]),_:2},1024))]}),_:2},1024)]),_:2},[s.notifications.length>0?{name:"notifications",fn:o(()=>[u("ul",null,[(i(!0),g(z,null,V(s.notifications,e=>(i(),g("li",{key:e.kind,"data-testid":`warning-${e.kind}`,innerHTML:l(`common.warnings.${e.kind}`,e.payload)},null,8,F))),128)),c()])]),key:"0"}:void 0]),1024)]),_:1})}}});export{v as default};
