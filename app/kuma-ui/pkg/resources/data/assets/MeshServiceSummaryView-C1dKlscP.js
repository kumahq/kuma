import{_ as F}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-BGUlZBqS.js";import{d as K,e as r,o as l,m as d,w as e,a as n,k as m,b as a,Y as u,t as c,p as _,c as x,H as C,J as z}from"./index-bM6gVJZj.js";import"./CodeBlock-DTe0mrry.js";const T={class:"stack"},A={class:"stack-with-borders"},M={class:"mt-4"},H=K({__name:"MeshServiceSummaryView",props:{items:{}},setup(b){const k=b;return(N,$)=>{const w=r("RouteTitle"),h=r("XAction"),g=r("KBadge"),S=r("KumaPort"),y=r("KTruncate"),R=r("DataSource"),V=r("AppView"),E=r("DataCollection"),P=r("RouteView");return l(),d(P,{name:"mesh-service-summary-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:e(({route:s,t:f,can:D})=>[n(E,{items:k.items,predicate:o=>o.id===s.params.service},{item:e(({item:o})=>[n(V,null,{title:e(()=>[m("h2",null,[n(h,{to:{name:"mesh-service-detail-view",params:{mesh:s.params.mesh,service:s.params.service}}},{default:e(()=>[n(w,{title:f("services.routes.item.title",{name:o.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:e(()=>[a(),m("div",T,[m("div",A,[n(u,{layout:"horizontal"},{title:e(()=>[a(`
                  State
                `)]),body:e(()=>[n(g,{appearance:o.spec.state==="Available"?"success":"danger"},{default:e(()=>[a(c(o.spec.state),1)]),_:2},1032,["appearance"])]),_:2},1024),a(),n(u,{layout:"horizontal"},{title:e(()=>[a(`
                  Dataplane Proxies
                `)]),body:e(()=>{var t,i,p;return[a(c((t=o.status.dataplaneProxies)==null?void 0:t.connected)+" connected, "+c((i=o.status.dataplaneProxies)==null?void 0:i.healthy)+" healthy ("+c((p=o.status.dataplaneProxies)==null?void 0:p.total)+` total)
                `,1)]}),_:2},1024),a(),o.namespace?(l(),d(u,{key:0,layout:"horizontal"},{title:e(()=>[a(`
                  Namespace
                `)]),body:e(()=>[a(c(o.namespace),1)]),_:2},1024)):_("",!0),a(),D("use zones")&&o.zone?(l(),d(u,{key:1,layout:"horizontal"},{title:e(()=>[a(`
                  Zone
                `)]),body:e(()=>[n(h,{to:{name:"zone-cp-detail-view",params:{zone:o.zone}}},{default:e(()=>[a(c(o.zone),1)]),_:2},1032,["to"])]),_:2},1024)):_("",!0),a(),n(u,{layout:"horizontal"},{title:e(()=>[a(`
                  Ports
                `)]),body:e(()=>[n(y,null,{default:e(()=>[(l(!0),x(C,null,z(o.spec.ports,t=>(l(),d(S,{key:t.port,port:{...t,targetPort:void 0}},null,8,["port"]))),128))]),_:2},1024)]),_:2},1024),a(),n(u,{layout:"horizontal"},{title:e(()=>[a(`
                  Selector
                `)]),body:e(()=>[n(y,null,{default:e(()=>[(l(!0),x(C,null,z(o.spec.selector.dataplaneTags,(t,i)=>(l(),d(g,{key:`${i}:${t}`,appearance:"info"},{default:e(()=>[a(c(i)+":"+c(t),1)]),_:2},1024))),128))]),_:2},1024)]),_:2},1024)]),a(),m("div",null,[m("h3",null,c(f("services.routes.item.config")),1),a(),m("div",M,[n(F,{resource:o.config,"is-searchable":"",query:s.params.codeSearch,"is-filter-mode":s.params.codeFilter,"is-reg-exp-mode":s.params.codeRegExp,onQueryChange:t=>s.update({codeSearch:t}),onFilterModeChange:t=>s.update({codeFilter:t}),onRegExpModeChange:t=>s.update({codeRegExp:t})},{default:e(({copy:t,copying:i})=>[i?(l(),d(R,{key:0,src:`/meshes/${s.params.mesh}/mesh-service/${s.params.service}/as/kubernetes?no-store`,onChange:p=>{t(v=>v(p))},onError:p=>{t((v,B)=>B(p))}},null,8,["src","onChange","onError"])):_("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])])]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1})}}});export{H as default};
