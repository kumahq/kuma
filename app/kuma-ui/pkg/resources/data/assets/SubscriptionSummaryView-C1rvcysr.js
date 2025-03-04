import{d as F,r as y,o as r,m as g,w as t,b as d,s as f,t as n,e,c as u,F as c,v as b,T as R,U as m,q as h,p as C,as as T,a as $,a2 as B}from"./index-ChH5weWG.js";const D={class:"stack-with-borders"},E={key:1,class:"mt-8 stack-with-borders"},M=F({__name:"SubscriptionSummaryView",props:{data:{},routeName:{}},setup(w){const _=w;return(S,s)=>{const I=y("XSelect"),k=y("XLayout"),v=y("XAlert"),x=y("XCodeBlock"),z=y("AppView"),V=y("DataCollection"),X=y("RouteView");return r(),g(X,{name:_.routeName,params:{subscription:"",codeSearch:"",codeFilter:!1,codeRegExp:!1,output:String}},{default:t(({route:p,t:l})=>[d(V,{items:_.data,predicate:a=>a.id===p.params.subscription},{item:t(({item:a})=>[d(z,null,{title:t(()=>[f("h2",null,n(a.zoneInstanceId??a.globalInstanceId??a.controlPlaneInstanceId),1)]),default:t(()=>[s[17]||(s[17]=e()),d(k,{type:"stack"},{default:t(()=>[f("header",null,[d(k,{type:"separated",justify:"end"},{default:t(()=>[(r(),u(c,null,b([["structured","yaml"]],o=>d(I,{key:typeof o,label:l("subscriptions.routes.item.format"),selected:p.params.output,onChange:i=>{p.update({output:i})}},R({_:2},[b(o,i=>({name:`${i}-option`,fn:t(()=>[e(n(l(`subscriptions.routes.item.formats.${i}`)),1)])}))]),1032,["label","selected","onChange"])),64))]),_:2},1024)]),s[16]||(s[16]=e()),p.params.output==="structured"?(r(),g(k,{key:0,type:"stack","data-testid":"structured-view"},{default:t(()=>[f("div",D,[d(m,{layout:"horizontal"},{title:t(()=>[e(n(l("http.api.property.version")),1)]),body:t(()=>{var o,i;return[(r(!0),u(c,null,b([(i=(o=a.version)==null?void 0:o.kumaCp)==null?void 0:i.version],A=>(r(),u(c,null,[e(n(A??"-"),1)],64))),256))]}),_:2},1024),s[6]||(s[6]=e()),d(m,{layout:"horizontal"},{title:t(()=>[e(n(l("http.api.property.connectTime")),1)]),body:t(()=>[e(n(l("common.formats.datetime",{value:Date.parse(a.connectTime??"")})),1)]),_:2},1024),s[7]||(s[7]=e()),a.disconnectTime?(r(),g(m,{key:0,layout:"horizontal"},{title:t(()=>[e(n(l("http.api.property.disconnectTime")),1)]),body:t(()=>[e(n(l("common.formats.datetime",{value:Date.parse(a.disconnectTime)})),1)]),_:2},1024)):h("",!0),s[8]||(s[8]=e()),d(m,{layout:"horizontal"},{title:t(()=>[e(n(l("subscriptions.routes.item.headers.responses")),1)]),body:t(()=>{var o;return[(r(!0),u(c,null,b([((o=a.status)==null?void 0:o.total)??{}],i=>(r(),u(c,null,[e(n(i.responsesSent)+"/"+n(i.responsesAcknowledged),1)],64))),256))]}),_:2},1024),s[9]||(s[9]=e()),(r(),u(c,null,b(["zoneInstanceId","globalInstanceId","controlPlaneInstanceId"],o=>(r(),u(c,{key:typeof o},[a[o]?(r(),g(m,{key:0,layout:"horizontal"},{title:t(()=>[e(n(l(`http.api.property.${o}`)),1)]),body:t(()=>[e(n(a[o]),1)]),_:2},1024)):h("",!0)],64))),64)),s[10]||(s[10]=e()),d(m,{layout:"horizontal"},{title:t(()=>[e(n(l("http.api.property.id")),1)]),body:t(()=>[e(n(a.id),1)]),_:2},1024)]),s[15]||(s[15]=e()),Object.keys(a.status.acknowledgements).length===0?(r(),g(v,{key:0,variant:"info"},{icon:t(()=>[d(C(T))]),default:t(()=>[e(" "+n(l("common.detail.subscriptions.no_stats",{id:a.id})),1)]),_:2},1024)):(r(),u("div",E,[f("div",null,[$(S.$slots,"default")]),s[13]||(s[13]=e()),d(m,{class:"mt-4",layout:"horizontal"},{title:t(()=>[f("strong",null,n(l("subscriptions.routes.item.headers.type")),1)]),body:t(()=>[e(n(l("subscriptions.routes.item.headers.stat")),1)]),_:2},1024),s[14]||(s[14]=e()),(r(!0),u(c,null,b(Object.entries(a.status.acknowledgements??{}),([o,i])=>(r(),g(m,{key:o,layout:"horizontal"},{title:t(()=>[e(n(l(`http.api.property.${o}`)),1)]),body:t(()=>[e(n(i.responsesSent)+"/"+n(i.responsesAcknowledged),1)]),_:2},1024))),128))]))]),_:2},1024)):(r(),g(x,{key:1,language:"yaml",code:C(B).stringify(a.$raw),"is-searchable":"",query:p.params.codeSearch,"is-filter-mode":p.params.codeFilter,"is-reg-exp-mode":p.params.codeRegExp,onQueryChange:o=>p.update({codeSearch:o}),onFilterModeChange:o=>p.update({codeFilter:o}),onRegExpModeChange:o=>p.update({codeRegExp:o})},null,8,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:3},8,["name"])}}});export{M as default};
