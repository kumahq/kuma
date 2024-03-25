import{C as A}from"./CodeBlock-QczJ3iSH.js";import{d as L,a as r,o,b as u,w as t,Z as d,f as e,e as a,t as _,W as g,q as h,m as n,F as y,c as i,H as k,p as N}from"./index-OvJp0Zog.js";import{T as S}from"./TagList--yHLjVWq.js";import{R as K}from"./RuleMatchers-q9CxnS4X.js";import{t as F}from"./toYaml-sPaYOD3i.js";const O={class:"stack-with-borders"},T={key:0,class:"mt-6"},H=n("h3",null,"Rules",-1),W={class:"stack mt-4"},j={class:"stack-with-borders"},q=n("dt",null,`
                        Config
                      `,-1),E={class:"mt-2"},Q=L({__name:"ConnectionInboundSummaryOverviewView",props:{data:{}},setup(w){const s=w;return(I,M)=>{const v=r("KBadge"),C=r("RouterLink"),x=r("DataSource"),B=r("KCard"),V=r("DataCollection"),z=r("DataLoader"),D=r("AppView"),P=r("RouteView");return o(),u(P,{params:{mesh:"",dataPlane:"",connection:""},name:"connection-inbound-summary-overview-view"},{default:t(({t:R,route:f})=>[a(D,null,{default:t(()=>[n("div",O,[a(d,{layout:"horizontal"},{title:t(()=>[e(`
            Tags
          `)]),body:t(()=>[a(S,{tags:s.data.tags,alignment:"right"},null,8,["tags"])]),_:1}),e(),a(d,{layout:"horizontal"},{title:t(()=>[e(`
            Status
          `)]),body:t(()=>[a(v,{appearance:s.data.health.ready?"success":"danger"},{default:t(()=>[e(_(s.data.health.ready?"Healthy":"Unhealthy"),1)]),_:1},8,["appearance"])]),_:1}),e(),a(d,{layout:"horizontal"},{title:t(()=>[e(`
            Protocol
          `)]),body:t(()=>[a(v,{appearance:"info"},{default:t(()=>[e(_(R(`http.api.value.${s.data.protocol}`)),1)]),_:2},1024)]),_:2},1024),e(),a(d,{layout:"horizontal"},{title:t(()=>[e(`
            Address
          `)]),body:t(()=>[a(g,{text:`${s.data.addressPort}`},null,8,["text"])]),_:1}),e(),s.data.serviceAddressPort.length>0?(o(),u(d,{key:0,layout:"horizontal"},{title:t(()=>[e(`
            Service Address
          `)]),body:t(()=>[a(g,{text:`${s.data.serviceAddressPort}`},null,8,["text"])]),_:1})):h("",!0)]),e(),s.data?(o(),i("div",T,[H,e(),a(z,{src:`/meshes/${f.params.mesh}/dataplanes/${f.params.dataPlane}/rules`},{default:t(({data:$})=>[a(V,{predicate:p=>p.ruleType==="from"&&Number(p.inbound.port)===Number(f.params.connection.split("_")[1]),items:$.rules},{default:t(({items:p})=>[n("div",W,[(o(!0),i(y,null,k(p,c=>(o(),u(B,{key:c},{default:t(()=>[n("div",j,[a(d,{layout:"horizontal"},{title:t(()=>[e(`
                        Type
                      `)]),body:t(()=>[e(_(c.type),1)]),_:2},1024),e(),c.matchers.length>0?(o(),u(d,{key:0,layout:"horizontal"},{title:t(()=>[e(`
                        From
                      `)]),body:t(()=>[n("p",null,[a(K,{items:c.matchers},null,8,["items"])])]),_:2},1024)):h("",!0),e(),c.origins.length>0?(o(),u(d,{key:1,layout:"horizontal"},{title:t(()=>[e(`
                        Origin Policies
                      `)]),body:t(()=>[a(x,{src:"/*/policy-types"},{default:t(({data:b})=>[(o(!0),i(y,null,k([Object.groupBy((b==null?void 0:b.policies)??[],m=>m.name)],m=>(o(),i("ul",{key:m},[(o(!0),i(y,null,k(c.origins,l=>(o(),i("li",{key:`${l.mesh}-${l.name}`},[m[l.type]?(o(),u(C,{key:0,to:{name:"policy-detail-view",params:{mesh:l.mesh,policyPath:m[l.type][0].path,policy:l.name}}},{default:t(()=>[e(_(l.name),1)]),_:2},1032,["to"])):(o(),i(y,{key:1},[e(_(l.name),1)],64))]))),128))]))),128))]),_:2},1024)]),_:2},1024)):h("",!0),e(),n("div",null,[q,e(),n("dd",E,[n("div",null,[a(A,{code:N(F)(c.config),language:"yaml","show-copy-button":!1},null,8,["code"])])])])])]),_:2},1024))),128))])]),_:2},1032,["predicate","items"])]),_:2},1032,["src"])])):h("",!0)]),_:2},1024)]),_:1})}}});export{Q as default};
