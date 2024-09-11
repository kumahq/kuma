import{_ as z}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-gTq31dOV.js";import{d as B,r as l,o as r,m,w as t,b as s,k as i,e as o,Z as h,c as d,L as f,M as g,t as c,p as k}from"./index-DW4c3gZM.js";import"./CodeBlock-C9q9QC_e.js";const D={class:"stack"},F={class:"stack-with-borders"},A={class:"mt-4"},q=B({__name:"MeshMultiZoneServiceSummaryView",props:{items:{}},setup(x){const w=x;return(P,T)=>{const b=l("RouteTitle"),R=l("XAction"),u=l("KTruncate"),v=l("KBadge"),S=l("DataSource"),V=l("AppView"),E=l("DataCollection"),M=l("RouteView");return r(),m(M,{name:"mesh-multi-zone-service-summary-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:t(({route:a,t:y})=>[s(E,{items:w.items,predicate:n=>n.id===a.params.service},{item:t(({item:n})=>[s(V,null,{title:t(()=>[i("h2",null,[s(R,{to:{name:"mesh-multi-zone-service-detail-view",params:{mesh:a.params.mesh,service:a.params.service}}},{default:t(()=>[s(b,{title:y("services.routes.item.title",{name:n.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:t(()=>[o(),i("div",D,[i("div",F,[n.status.addresses.length>0?(r(),m(h,{key:0,layout:"horizontal"},{title:t(()=>[o(`
                  Addresses
                `)]),body:t(()=>[s(u,null,{default:t(()=>[(r(!0),d(f,null,g(n.status.addresses,e=>(r(),d("span",{key:e.hostname},c(e.hostname),1))),128))]),_:2},1024)]),_:2},1024)):k("",!0),o(),s(h,{layout:"horizontal"},{title:t(()=>[o(`
                  Ports
                `)]),body:t(()=>[s(u,null,{default:t(()=>[(r(!0),d(f,null,g(n.spec.ports,e=>(r(),m(v,{key:e.port,appearance:"info"},{default:t(()=>[o(c(e.port)+c(e.targetPort?`:${e.targetPort}`:"")+c(e.appProtocol?`/${e.appProtocol}`:""),1)]),_:2},1024))),128))]),_:2},1024)]),_:2},1024),o(),s(h,{layout:"horizontal"},{title:t(()=>[o(`
                  Match Labels
                `)]),body:t(()=>[s(u,null,{default:t(()=>[(r(!0),d(f,null,g(n.spec.selector.meshService.matchLabels,(e,p)=>(r(),m(v,{key:`${p}:${e}`,appearance:"info"},{default:t(()=>[o(c(p)+":"+c(e),1)]),_:2},1024))),128))]),_:2},1024)]),_:2},1024)]),o(),i("div",null,[i("h3",null,c(y("services.routes.item.config")),1),o(),i("div",A,[s(z,{resource:n.config,"is-searchable":"",query:a.params.codeSearch,"is-filter-mode":a.params.codeFilter,"is-reg-exp-mode":a.params.codeRegExp,onQueryChange:e=>a.update({codeSearch:e}),onFilterModeChange:e=>a.update({codeFilter:e}),onRegExpModeChange:e=>a.update({codeRegExp:e})},{default:t(({copy:e,copying:p})=>[p?(r(),m(S,{key:0,src:`/meshes/${a.params.mesh}/mesh-multi-zone-service/${a.params.service}/as/kubernetes?no-store`,onChange:_=>{e(C=>C(_))},onError:_=>{e((C,$)=>$(_))}},null,8,["src","onChange","onError"])):k("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])])]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1})}}});export{q as default};
