import{_ as F}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-BGUlZBqS.js";import{d as K,e as a,o as c,m,w as t,a as s,k as i,b as n,Y as g,c as f,H as v,J as C,t as d,p as M}from"./index-bM6gVJZj.js";import"./CodeBlock-DTe0mrry.js";const z={class:"stack"},T={class:"stack-with-borders"},$={class:"mt-4"},Q=K({__name:"MeshMultiZoneServiceSummaryView",props:{items:{}},setup(y){const x=y;return(A,N)=>{const w=a("RouteTitle"),S=a("XAction"),k=a("KumaPort"),u=a("KTruncate"),R=a("KBadge"),V=a("DataSource"),E=a("AppView"),b=a("DataCollection"),B=a("RouteView");return c(),m(B,{name:"mesh-multi-zone-service-summary-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:t(({route:o,t:_})=>[s(b,{items:x.items,predicate:r=>r.id===o.params.service},{item:t(({item:r})=>[s(E,null,{title:t(()=>[i("h2",null,[s(S,{to:{name:"mesh-multi-zone-service-detail-view",params:{mesh:o.params.mesh,service:o.params.service}}},{default:t(()=>[s(w,{title:_("services.routes.item.title",{name:r.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:t(()=>[n(),i("div",z,[i("div",T,[s(g,{layout:"horizontal"},{title:t(()=>[n(`
                  Ports
                `)]),body:t(()=>[s(u,null,{default:t(()=>[(c(!0),f(v,null,C(r.spec.ports,e=>(c(),m(k,{key:e.port,port:{...e,targetPort:void 0}},null,8,["port"]))),128))]),_:2},1024)]),_:2},1024),n(),s(g,{layout:"horizontal"},{title:t(()=>[n(`
                  Selector
                `)]),body:t(()=>[s(u,null,{default:t(()=>[(c(!0),f(v,null,C(r.spec.selector.meshService.matchLabels,(e,l)=>(c(),m(R,{key:`${l}:${e}`,appearance:"info"},{default:t(()=>[n(d(l)+":"+d(e),1)]),_:2},1024))),128))]),_:2},1024)]),_:2},1024)]),n(),i("div",null,[i("h3",null,d(_("services.routes.item.config")),1),n(),i("div",$,[s(F,{resource:r.config,"is-searchable":"",query:o.params.codeSearch,"is-filter-mode":o.params.codeFilter,"is-reg-exp-mode":o.params.codeRegExp,onQueryChange:e=>o.update({codeSearch:e}),onFilterModeChange:e=>o.update({codeFilter:e}),onRegExpModeChange:e=>o.update({codeRegExp:e})},{default:t(({copy:e,copying:l})=>[l?(c(),m(V,{key:0,src:`/meshes/${o.params.mesh}/mesh-multi-zone-service/${o.params.service}/as/kubernetes?no-store`,onChange:p=>{e(h=>h(p))},onError:p=>{e((h,D)=>D(p))}},null,8,["src","onChange","onError"])):M("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])])]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1})}}});export{Q as default};
