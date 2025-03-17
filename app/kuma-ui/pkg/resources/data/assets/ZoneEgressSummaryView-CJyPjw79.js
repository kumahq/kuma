import{d as D,r as d,o as m,m as c,w as t,b as r,s as u,t as i,e as s,c as f,F as _,v as x,T as q,U as h,S as L,q as T,p as N,ar as Q,_ as Z}from"./index-Bi3CXAeE.js";import{_ as C}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-SewON5TQ.js";const v=D({__name:"ZoneEgressSummaryView",props:{items:{}},setup(E){const z=E;return(I,a)=>{const S=d("XEmptyState"),b=d("RouteTitle"),w=d("XAction"),R=d("XSelect"),y=d("XLayout"),V=d("XCopyButton"),F=d("DataLoader"),M=d("AppView"),X=d("DataCollection"),B=d("RouteView");return m(),c(B,{name:"zone-egress-summary-view",params:{proxy:"",codeSearch:"",codeFilter:!1,codeRegExp:!1,format:String}},{default:t(({route:e,t:l,uri:A})=>[r(X,{items:z.items,predicate:g=>g.id===e.params.proxy,find:!0},{empty:t(()=>[r(S,null,{title:t(()=>[u("h2",null,i(l("common.collection.summary.empty_title",{type:"ZoneEgress"})),1)]),default:t(()=>[a[0]||(a[0]=s()),u("p",null,i(l("common.collection.summary.empty_message",{type:"ZoneEgress"})),1)]),_:2},1024)]),default:t(({items:g})=>[(m(!0),f(_,null,x([g[0]],p=>(m(),c(M,{key:p.id},{title:t(()=>[u("h2",null,[r(w,{to:{name:"zone-egress-detail-view",params:{zone:p.zoneEgress.zone,proxyType:"egresses",proxy:p.id}}},{default:t(()=>[r(b,{title:l("zone-egresses.routes.item.title",{name:p.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:t(()=>[a[8]||(a[8]=s()),r(y,{type:"stack"},{default:t(()=>[u("header",null,[r(y,{type:"separated",size:"max"},{default:t(()=>[u("h3",null,i(l("zone-ingresses.routes.item.config")),1),a[1]||(a[1]=s()),(m(),f(_,null,x([["structured","universal","k8s"]],n=>u("div",{key:typeof n},[r(R,{label:l("zone-ingresses.routes.items.format"),selected:e.params.format,onChange:o=>{e.update({format:o})},onVnodeBeforeMount:o=>{var k;return((k=o==null?void 0:o.props)==null?void 0:k.selected)&&n.includes(o.props.selected)&&o.props.selected!==e.params.format&&e.update({format:o.props.selected})}},q({_:2},[x(n,o=>({name:`${o}-option`,fn:t(()=>[s(i(l(`zone-egresses.routes.items.formats.${o}`)),1)])}))]),1032,["label","selected","onChange","onVnodeBeforeMount"])])),64))]),_:2},1024)]),a[7]||(a[7]=s()),e.params.format==="structured"?(m(),c(y,{key:0,type:"stack",class:"stack-with-borders","data-testid":"structured-view"},{default:t(()=>[r(h,{layout:"horizontal"},{title:t(()=>[s(i(l("http.api.property.status")),1)]),body:t(()=>[r(L,{status:p.state},null,8,["status"])]),_:2},1024),a[5]||(a[5]=s()),p.namespace.length>0?(m(),c(h,{key:0,layout:"horizontal"},{title:t(()=>[s(i(l("data-planes.routes.item.namespace")),1)]),body:t(()=>[s(i(p.namespace),1)]),_:2},1024)):T("",!0),a[6]||(a[6]=s()),r(h,{layout:"horizontal"},{title:t(()=>[s(i(l("http.api.property.address")),1)]),body:t(()=>[p.zoneEgress.socketAddress.length>0?(m(),c(V,{key:0,text:p.zoneEgress.socketAddress},null,8,["text"])):(m(),f(_,{key:1},[s(i(l("common.detail.none")),1)],64))]),_:2},1024)]),_:2},1024)):e.params.format==="universal"?(m(),c(C,{key:1,"data-testid":"codeblock-yaml-universal",language:"yaml",resource:p.config,"show-k8s-copy-button":!1,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:n=>e.update({codeSearch:n}),onFilterModeChange:n=>e.update({codeFilter:n}),onRegExpModeChange:n=>e.update({codeRegExp:n})},null,8,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])):(m(),c(F,{key:2,src:A(N(Q),"/zone-egresses/:name/as/kubernetes",{name:e.params.proxy})},{default:t(({data:n})=>[r(C,{"data-testid":"codeblock-yaml-k8s",language:"yaml",resource:n,"show-k8s-copy-button":!1,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},null,8,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"]))]),_:2},1024)]),_:2},1024))),128))]),_:2},1032,["items","predicate"])]),_:1})}}}),G=Z(v,[["__scopeId","data-v-775d9346"]]);export{G as default};
