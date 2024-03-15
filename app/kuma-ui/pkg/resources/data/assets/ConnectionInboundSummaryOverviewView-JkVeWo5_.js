import{C as N}from"./CodeBlock-bFIOzW2Q.js";import{d as S,a as i,o,b as m,w as t,Z as s,f as e,e as a,c as l,t as _,W as v,q as h,F as y,m as r,H as g,p as A}from"./index-RCz0LtdB.js";import{T as w}from"./TagList-eJZp0YaP.js";import{R as K}from"./RuleMatchers-_Kp0uSD2.js";import{t as T}from"./toYaml-sPaYOD3i.js";const F={class:"stack"},O={key:0,class:"stack-with-borders"},H={key:1,class:"stack-with-borders"},W={key:2},j=r("h3",null,"Rules",-1),q={class:"stack mt-4"},E={class:"stack-with-borders mt-4"},I=r("dt",null,`
                          Config
                        `,-1),M={class:"mt-2"},tt=S({__name:"ConnectionInboundSummaryOverviewView",props:{inbound:{},gateway:{}},setup(C){const n=C;return(U,Y)=>{const k=i("KBadge"),x=i("RouterLink"),z=i("DataSource"),B=i("KCard"),V=i("DataCollection"),D=i("DataLoader"),P=i("AppView"),R=i("RouteView");return o(),m(R,{name:"connection-inbound-summary-overview-view",params:{mesh:"",dataPlane:"",service:""}},{default:t(({route:b,t:$})=>[a(P,null,{default:t(()=>[r("div",F,[n.gateway?(o(),l("div",O,[a(s,{layout:"horizontal"},{title:t(()=>[e(`
              Tags
            `)]),body:t(()=>[a(w,{tags:n.gateway.tags,alignment:"right"},null,8,["tags"])]),_:1})])):n.inbound?(o(),l("div",H,[a(s,{layout:"horizontal"},{title:t(()=>[e(`
              Tags
            `)]),body:t(()=>[a(w,{tags:n.inbound.tags,alignment:"right"},null,8,["tags"])]),_:1}),e(),a(s,{layout:"horizontal"},{title:t(()=>[e(`
              Status
            `)]),body:t(()=>[a(k,{appearance:n.inbound.health.ready?"success":"danger"},{default:t(()=>[e(_(n.inbound.health.ready?"Healthy":"Unhealthy"),1)]),_:1},8,["appearance"])]),_:1}),e(),a(s,{layout:"horizontal"},{title:t(()=>[e(`
              Protocol
            `)]),body:t(()=>[a(k,{appearance:"info"},{default:t(()=>[e(_($(`http.api.value.${n.inbound.protocol}`)),1)]),_:2},1024)]),_:2},1024),e(),a(s,{layout:"horizontal"},{title:t(()=>[e(`
              Address
            `)]),body:t(()=>[a(v,{text:`${n.inbound.addressPort}`},null,8,["text"])]),_:1}),e(),a(s,{layout:"horizontal"},{title:t(()=>[e(`
              Service Address
            `)]),body:t(()=>[a(v,{text:`${n.inbound.serviceAddressPort}`},null,8,["text"])]),_:1})])):h("",!0),e(),n.inbound?(o(),l("div",W,[j,e(),a(D,{src:`/meshes/${b.params.mesh}/dataplanes/${b.params.dataPlane}/rules`},{default:t(({data:L})=>[a(V,{predicate:p=>p.ruleType==="from"&&Number(p.inbound.port)===Number(b.params.service.substring(1)),items:L.rules},{default:t(({items:p})=>[r("div",q,[(o(!0),l(y,null,g(p,c=>(o(),m(B,{key:c},{default:t(()=>[r("div",E,[a(s,{class:"mt-2",layout:"horizontal"},{title:t(()=>[e(`
                          Type
                        `)]),body:t(()=>[e(_(c.type),1)]),_:2},1024),e(),c.matchers.length>0?(o(),m(s,{key:0,layout:"horizontal"},{title:t(()=>[e(`
                          From
                        `)]),body:t(()=>[r("p",null,[a(K,{items:c.matchers},null,8,["items"])])]),_:2},1024)):h("",!0),e(),c.origins.length>0?(o(),m(s,{key:1,layout:"horizontal"},{title:t(()=>[e(`
                          Origin Policies
                        `)]),body:t(()=>[a(z,{src:"/*/policy-types"},{default:t(({data:f})=>[(o(!0),l(y,null,g([Object.groupBy((f==null?void 0:f.policies)??[],u=>u.name)],u=>(o(),l("ul",{key:u},[(o(!0),l(y,null,g(c.origins,d=>(o(),l("li",{key:`${d.mesh}-${d.name}`},[u[d.type]?(o(),m(x,{key:0,to:{name:"policy-detail-view",params:{mesh:d.mesh,policyPath:u[d.type][0].path,policy:d.name}}},{default:t(()=>[e(_(d.name),1)]),_:2},1032,["to"])):(o(),l(y,{key:1},[e(_(d.name),1)],64))]))),128))]))),128))]),_:2},1024)]),_:2},1024)):h("",!0),e(),r("div",null,[I,e(),r("dd",M,[r("div",null,[a(N,{code:A(T)(c.config),language:"yaml","show-copy-button":!1},null,8,["code"])])])])])]),_:2},1024))),128))])]),_:2},1032,["predicate","items"])]),_:2},1032,["src"])])):h("",!0)])]),_:2},1024)]),_:1})}}});export{tt as default};
