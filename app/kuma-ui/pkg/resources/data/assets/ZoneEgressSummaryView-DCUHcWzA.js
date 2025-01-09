import{d as D,r as l,o as d,p as c,w as e,b as i,l as m,t as p,e as t,c as f,J as E,K as F,Q as _,S as X,q as z,_ as M}from"./index-BIN9nSPF.js";import{_ as N}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-3fFCInp0.js";const Z={class:"stack-with-borders"},q={class:"mt-4"},Q=D({__name:"ZoneEgressSummaryView",props:{items:{}},setup(C){const h=C;return(T,o)=>{const x=l("XEmptyState"),k=l("RouteTitle"),S=l("XAction"),w=l("XCopyButton"),V=l("DataSource"),R=l("AppView"),v=l("DataCollection"),b=l("RouteView");return d(),c(b,{name:"zone-egress-summary-view",params:{zoneEgress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:e(({route:n,t:a})=>[i(v,{items:h.items,predicate:u=>u.id===n.params.zoneEgress,find:!0},{empty:e(()=>[i(x,null,{title:e(()=>[m("h2",null,p(a("common.collection.summary.empty_title",{type:"ZoneEgress"})),1)]),default:e(()=>[o[0]||(o[0]=t()),m("p",null,p(a("common.collection.summary.empty_message",{type:"ZoneEgress"})),1)]),_:2},1024)]),default:e(({items:u})=>[(d(!0),f(E,null,F([u[0]],s=>(d(),c(R,{key:s.id},{title:e(()=>[m("h2",null,[i(S,{to:{name:"zone-egress-detail-view",params:{zone:s.zoneEgress.zone,zoneEgress:s.id}}},{default:e(()=>[i(k,{title:a("zone-egresses.routes.item.title",{name:s.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:e(()=>[o[7]||(o[7]=t()),m("div",Z,[i(_,{layout:"horizontal"},{title:e(()=>[t(p(a("http.api.property.status")),1)]),body:e(()=>[i(X,{status:s.state},null,8,["status"])]),_:2},1024),o[4]||(o[4]=t()),s.namespace.length>0?(d(),c(_,{key:0,layout:"horizontal"},{title:e(()=>[t(p(a("data-planes.routes.item.namespace")),1)]),body:e(()=>[t(p(s.namespace),1)]),_:2},1024)):z("",!0),o[5]||(o[5]=t()),i(_,{layout:"horizontal"},{title:e(()=>[t(p(a("http.api.property.address")),1)]),body:e(()=>[s.zoneEgress.socketAddress.length>0?(d(),c(w,{key:0,text:s.zoneEgress.socketAddress},null,8,["text"])):(d(),f(E,{key:1},[t(p(a("common.detail.none")),1)],64))]),_:2},1024)]),o[8]||(o[8]=t()),m("div",null,[m("h3",null,p(a("zone-ingresses.routes.item.config")),1),o[6]||(o[6]=t()),m("div",q,[i(N,{resource:s.config,"is-searchable":"",query:n.params.codeSearch,"is-filter-mode":n.params.codeFilter,"is-reg-exp-mode":n.params.codeRegExp,onQueryChange:r=>n.update({codeSearch:r}),onFilterModeChange:r=>n.update({codeFilter:r}),onRegExpModeChange:r=>n.update({codeRegExp:r})},{default:e(({copy:r,copying:B})=>[B?(d(),c(V,{key:0,src:`/zone-egresses/${n.params.zoneEgress}/as/kubernetes?no-store`,onChange:g=>{r(y=>y(g))},onError:g=>{r((y,A)=>A(g))}},null,8,["src","onChange","onError"])):z("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])]),_:2},1024))),128))]),_:2},1032,["items","predicate"])]),_:1})}}}),J=M(Q,[["__scopeId","data-v-35c949ce"]]);export{J as default};
