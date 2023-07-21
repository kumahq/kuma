import{d as Y,r as J,o as e,e as a,g as s,F as f,j as O,k as T,t as v,h as u,w as t,f as L,a as d,l as Pe,b as z,p as he,m as ge,c as x,n as ke,q as I,s as Ee,K as Oe,u as Qe}from"./index-f4ec0be6.js";import{N as Ge,u as Ue,g as fe,K as Ie}from"./kongponents.es-fed304fd.js";import{A as W,a as V,_ as Me,S as Le}from"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-a76df740.js";import{f as N,e as te,g as ae,h as ve,i as Te,j as ze,k as ne,l as Re,p as xe,m as Se,C as Ye,I as Ne,n as _e,A as He,_ as qe}from"./RouteView.vue_vue_type_script_setup_true_lang-09fd82a0.js";import{_ as De}from"./CodeBlock.vue_vue_type_style_index_0_lang-56242a2b.js";import{T as j}from"./TagList-0012d9cb.js";import{_ as we,T as Ke}from"./TextWithCopyButton-ae3a8132.js";import{E as Be}from"./ErrorBlock-f3ed6714.js";import{t as le,_ as je}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-7c8e38a7.js";import{D as ie,a as K}from"./DefinitionListItem-0b3f80a7.js";import{E as ee}from"./EnvoyData-c3ce46be.js";import{S as Fe}from"./StatusBadge-0a8731e8.js";import{_ as Je}from"./StatusInfo.vue_vue_type_script_setup_true_lang-7a761e1d.js";import{T as We}from"./TabsWidget-2ca8d57d.js";import{_ as Ve}from"./WarningsWidget.vue_vue_type_script_setup_true_lang-eefa3964.js";import{_ as Xe}from"./RouteTitle.vue_vue_type_script_setup_true_lang-d39e0ee5.js";import"./QueryParameter-70743f73.js";const q=o=>(he("data-v-1a7d780a"),o=o(),ge(),o),Ze={class:"mesh-gateway-policy-list"},$e=q(()=>T("h3",null,"Gateway policies",-1)),et={key:0,class:"policy-list"},tt=q(()=>T("h3",{class:"mt-6"},`
      Listeners
    `,-1)),at=q(()=>T("b",null,"Host",-1)),st=q(()=>T("h4",{class:"mt-2"},`
              Routes
            `,-1)),nt={class:"dataplane-policy-header"},lt=q(()=>T("b",null,"Route",-1)),it=q(()=>T("b",null,"Service",-1)),ot={key:0,class:"badge-list"},At={class:"policy-list mt-1"},ct=Y({__name:"MeshGatewayDataplanePolicyList",props:{meshGatewayDataplane:{type:Object,required:!0},meshGatewayListenerEntries:{type:Array,required:!0},meshGatewayRoutePolicies:{type:Array,required:!0}},setup(o){const l=o;return(B,C)=>{const D=J("router-link");return e(),a("div",Ze,[$e,s(),o.meshGatewayRoutePolicies.length>0?(e(),a("ul",et,[(e(!0),a(f,null,O(o.meshGatewayRoutePolicies,(y,h)=>(e(),a("li",{key:h},[T("span",null,v(y.type),1),s(`:

        `),u(D,{to:y.route},{default:t(()=>[s(v(y.name),1)]),_:2},1032,["to"])]))),128))])):L("",!0),s(),tt,s(),T("div",null,[(e(!0),a(f,null,O(l.meshGatewayListenerEntries,(y,h)=>(e(),a("div",{key:h},[T("div",null,[T("div",null,[at,s(": "+v(y.hostName)+":"+v(y.port)+" ("+v(y.protocol)+`)
          `,1)]),s(),y.routeEntries.length>0?(e(),a(f,{key:0},[st,s(),u(V,{"initially-open":[],"multiple-open":""},{default:t(()=>[(e(!0),a(f,null,O(y.routeEntries,(c,g)=>(e(),d(W,{key:g},Pe({"accordion-header":t(()=>[T("div",nt,[T("div",null,[T("div",null,[lt,s(": "),u(D,{to:c.route},{default:t(()=>[s(v(c.routeName),1)]),_:2},1032,["to"])]),s(),T("div",null,[it,s(": "+v(c.service),1)])]),s(),c.policies.length>0?(e(),a("div",ot,[(e(!0),a(f,null,O(c.policies,(n,m)=>(e(),d(z(Ge),{key:`${h}-${m}`},{default:t(()=>[s(v(n.type),1)]),_:2},1024))),128))])):L("",!0)])]),_:2},[c.policies.length>0?{name:"accordion-content",fn:t(()=>[T("ul",At,[(e(!0),a(f,null,O(c.policies,(n,m)=>(e(),a("li",{key:`${h}-${m}`},[s(v(n.type)+`:

                      `,1),u(D,{to:n.route},{default:t(()=>[s(v(n.name),1)]),_:2},1032,["to"])]))),128))])]),key:"0"}:void 0]),1024))),128))]),_:2},1024)],64)):L("",!0)])]))),128))])])}}});const rt=N(ct,[["__scopeId","data-v-1a7d780a"]]),oe="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAHgAAAB4CAMAAAAOusbgAAAAVFBMVEXa2tra2tra2tra2tra2tra2tr////a2toAfd6izPLvzPnRfvDYteSKr86zas0Aar4AhODY6vr3+Prx8v2Kv+9aqOk3muUOj+N5t+211vXhqfW01fXvn55GAAAABnRSTlMC9s/Hbhsvz/I3AAABVklEQVRo3u3b3Y6CMBCG4SJYhnV/KD+K7v3f57bN7AFJTcDUmZB+74lH5EmMA5hmjK+pq1awqm5M6HxqxTudPSzssmxM06rUmDp8DFawIYi1qYRdlisTeCtcMAGnAgwYMGDAgJ8GGPDB4B8frepnl9cZH5d1374E7GmX1WVuA0xzTvixA+5zwpc0/OXrVgU5N/yx6tMHGDBgwIABvxmeiBZhmF3fPMjDFLuOSjDdnBJMvVOAb1G+y8PjlUKdOGyHOcpLJniiDfEVC/FYZYA3unxFx2OVAd7sTjZ073msRGB2Yy7KvcsC2z05Hitx2P6PVTEwf9W/h/5xvTBOB76ByN8ydzRRzofELln1schjVNCrTxyjsl5vtV7ol7L+tAEGDLhMWOAw5ADHPxIHXmpHfAWepgJOBBgwYMCAAT8NMGDAgJOw2hKO2tqR2qKV1mqZ3jKd2vrgH/W3idgykdWgAAAAAElFTkSuQmCC",ut="data:image/png;base64,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",Ae="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAYAAABXAvmHAAAABGdBTUEAALGPC/xhBQAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAMKADAAQAAAABAAAAMAAAAAD4/042AAAEj0lEQVRoBe1aS28TVxT+7ngydhwnPGISTBKHEB6iUtOWHTvWqCtKqQhISC2vBbCpRDf9BUgsgAWbdlGppWqlSl1UXfMLCmXRqgXUxmCclOCWxI4Tv4Zz7s00BntmLh4rTCSfxJ4755458333fHfuTTQCZFOHTo+ijCs2cAi2nWJfaE2InABuw8Lle7e/eCwYvL2CXwF7a2hBtwQm8iKKdwwe+Y0HnhnRgBN2Q8qmJcPwOxm7EXrNe40jzVfDq38j9HUJvOkqdSvQrUDAEeiIhGaPH8bsyfe1oWQuTuPxhePa8V6BplenTl85tQ2l9A7YMUsnHMsTIyjtm9CK1QkKXIHC3nEI2l3RgqhzPzw/sB/g+A5ZYAKlPTsVFMnCH1Xx3f26XP2TUUQgAuXhJKr9fQqQRgVYPpUtA7IANvQq5sciEIHi7jHKb5OE9DQh5SOvoGs6pKNABJYn06tAaDQ1SLB82DoFnnO1TaA8NIhqIo7IQkFLDI58zPx/WvEMTsfaJlAiPbPF789oiWHxPTX6A3f/kPGdmQEBKlCaGJUE+oiANJ9JvEAEeOL23/ldHvVmjUrt9d1WBSrJLaiRfMzCEqzcU8pPcDzmAMunSk8f699FxP7KqngvVK/R19ZKvDy+Qy5cvQ8z8la2xuhzII8+m9foF9+axOz0YRm3/dbP6PvtoWy7fZm1iIV6tAd1i4+W3BLUrR7Y1Jb+1T7eKqg41ccajj94JPPy4DskaoleZM8cRYmeUGyO1hm0Q6DRz5XMnj2KpV1jTcSYyOTnNzjc1Uw1eCwBpQIFhNWqfvhKCZDPZbCQoGK5eVhz82uJKYjBPDp/DFwhBswZnEcmT3YlnzV/jRbBzKVplFNDTeDXEnu3TLNeBpb44x3o20vksh8fQYU2d1GaF+nr3yBCc6SVOaQyl05gxYm/9rWMf1VCra5v9LU1BxoT/N+mCpSHB2HNzmP05neu4J14ltZKKqnIroLnPta8n2ycHHzsHAGqgPXPM4x8+QOBLzXeo6ntSMsiGaYbwDcFajg6QiA6k0M9EQM/NSJFb/CMqe/PDD0QTKrU976V8uMg3j74ifOg8IsNZX9bC1mYmHQJvOlqBJ7EcUPgw8EELFq5vn1WQKHmPaX6IwIXhzdJ3jfmnmPRJ95vgAJJqJfAf0Tgx3pMpGn7cW5oExIE0M0Y/GepzdgT65EfbrPvVZuKW7g6vlV+uO1lYurgWTtmGHIEo7QYxYhSlM6jlJf9UT6nNvtiBFj5+SjUNeRbrNWpLTBmRSiOc6h8bjfOlquya8TyEQDdN1+t4dOZvFsqXsjU3ob/rqVfMv5iGaijbdORO2ihUlshiqdu5RZ4Uqnix3wRBsWcSiawj/8/xAEqGSd8ye4vV8DS4e3EheEBWYmXAl7zJJTrAMvm1LaEpPLV0wLu8V7NxUJJwAVrS3egSdwy4zo7uwTWecCbbtetQNOQrLPDoOd1bp3v2bnbEXZaN+nFiQ1qjJ3WfFymZdN9rQ4tOcJM2CNzf/+ysH33gVuiLlIkpyTh7Q8tZgbGr9sI8RO9qfIBv27zAiEVYZQrGIvuAAAAAElFTkSuQmCC",ce="data:image/png;base64,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",re="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAHgAAAB5CAYAAADyOOV3AAAFOklEQVR4Ae2dP2tUQRTFp7S385MofgRFiGBhKr9GuqBiI9iktwosCAnRLo0g8Q+ojSaNBomumESTIAqSLiPTTc4k7+bmztudu3sCAztv7p137/nNebtuREMIIXz9eXBluLO/NNzZe8sxCRrsL23tHlxObMP33b3ZzeHO0edv25FjcjRITBPbsPXj12+CnRywOcvENtC9kwk3gU5sQ048vf7775DDsQbIswAc+eNaAQJ2jU8unoBljVxHELBrfHLxBCxr5DqCgF3jk4snYFkj1xEE7BqfXDwByxq5jiBg1/jk4glY1sh1BAG7xicXT8CyRq4jCNg1Prl4ApY1ch1BwK7xycUTsKyR6wgCdo1PLp6AZY1cRxCwa3xy8QQsa+Q6goBd45OLJ2BZI9cRBOwan1w8AcsauY4gYNf45OIJWNbIdQQBu8YnF0/AskauIwjYNT65eAKWNXIdUQD+c2sm5iPemY2mIcnD/bsVMuqTs0yvQ7wQYtXRXb79XtxfpSEB4wH3foCgHwIGQSS+5qeddAOsxzgPOwsLMR9xsBhNQ2qA+3crZNQnZ5le89/o6Jbb3WrxKRovuOuIBR9TAHnSwcfk8T8hYP8MOzsg4E55/C8SsH+GnR0QcKc8/hcJ2D/Dzg4IuFMe/4sE7J9hZwcE3CmP/8WpAzz7cCnm48bdQaw58r3T63H/TB3gcG0+jnIQ8IgVGCXcdK9x/9DBPTuagEesADr43uBFrDlwf217+B5unV+fX4z5mPjfJiGA95vbsebA/bWAMb/6HJ/Z2gJbj0fBasJNe+H+Wj0wv/qcgG2ORiAErFXAGI8AWnfwo5U30TLmHq/GfPA92PiejAdIex4x33oAl9c+xnwQMAFrz2Rb8bUdgg7D/bXdYz7ur53n7k2v6WA6WHsm24qv7RB0FO6v7R7zcX/tnA42OhYFR0AErFXAGI8AEJB1jvtry8V8az2Fg/PvLdNr63ehmK9tuHZ8bQERAO6vrR/zcX/tvACMN6g91zZcOx770QomxeP+2voxX7qftE7A0/YejCeo9lx7omvHYz+SA7TruL+2fszX3h/jCwfn31um15bvQVMuFqxtGN/DrXOsBwWxznF/bb+Yb62nAIwXrDfAgq0N437WubU/zMd6rP3i/to58gx4QbshxtduGPezzrFe6xzrIWBBARSs9twKFPOxPqG9YhnzcX/tHA3bvIOtnwkwXyuYFI+ACoLCBcyX7ietuwMsNTTudQQk8CyWMd/aDwFP25+Dkbj1BOGJLI6scAHzrfX0nY/1Cu0Vy5hvrRd5Nv8ebG2473wEVBAULmC+tV4C5iO6rb9Gaj3RfeejAwXDFsuYb62XDqaD6WCNi9CBhUWFC5ivufdJsXQwHUwHn+SM066hAwXDFsuYf9p9znqdDqaD6zpY+/vc2if6rCf/vHFY77j7HbmDUQDt/LzCjypP248Ub62bgHt+REsApXUCrgzIKqgETLturad3B+PvX61za8N951v7w3xrvb0DthbIfNuHXAJu7BFf+0ATMAHbHhG1TyT30/Ggg+lg3Ymhw9rSiw6mg9s6kXxC6HjQwXSw7sTQYW3pRQfTwW2dSD4hdDzoYDpYd2LosLb0ooPp4LZOJJ8QOh50MB2sOzF0WFt60cF0cFsnkk8IHQ86mA7WnRg6rC296OBpd/Dqu0+Rw68GhYNXXq4f4UXOj//fQ171SGzD8tr60GsDrFs6iOvDcPP+k5mnrzYOKZYklq/1xDSxDWHmwcWr84NLz15v3H7+4csch38NEsvENLH9DwLs1co+Fv2iAAAAAElFTkSuQmCC",ue=""+new URL("Retry-8b2ec896.png",import.meta.url).href,pe=""+new URL("Timeout-dcabf0f7.jpg",import.meta.url).href,de="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAYAAABXAvmHAAAABGdBTUEAALGPC/xhBQAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAMKADAAQAAAABAAAAMAAAAAD4/042AAABYklEQVRoBe2av0oDQRDGZxbRxhfwDRI0NhKtRAhWPkM6Ex9KTOczWElArBRsAuEeIS+QRpvJfJdcqkWRLWYH5or7s7N797v59j4Odph2m4hw//xywsT3JHQqJMddrIajcq2Jaalcs2bx+cTMAi7Grn9xfSI/388kMsJ19RvznA+Pxs3X+yoh867gkV1NNJjBzr3BcKpT5rH6rOcAmR5SO+dzQQdtYE/4YB2w5hGVPdXmNnnSfCvYUz7kpzVewFor9woc/DeDb/OXX4fcjO728b/67jsWnLhXgHtnw/anqCAJpkPdKxAvYDp/9OHhQtYKhAtZKxAuZK1AuJC1AuFC1gqEC1krEC5krUC4kLUC4ULWCoQLWSsQLmStQLhQKFCYAaxSrgvvYTYc7AnL92YEpQ9WdqxSzkrvYzUe7Lwt8rh6dVMn0WVL6yWaxcdtQtUHCidIG7pY9cddsUfL3sF6LbfZAN5wf/+tIkpkAAAAAElFTkSuQmCC",ye="data:image/png;base64,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",pt="data:image/png;base64,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",me="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAYAAABXAvmHAAAABGdBTUEAALGPC/xhBQAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAMKADAAQAAAABAAAAMAAAAAD4/042AAAEs0lEQVRoBe1azWtUVxQ/781HkslHzQwpDfkQUUpdaHZtaUtTuhACYtC/wI22FHd+bMSlFNSdIhjcddFNKW0pZlfS0BayEdSFqAjRJMbGfBgTZ+JM8p7nd27ezMvkvsy7yUucAS+898479+P8zj3n3nvembGIy8xMttOhwiVy6RuX3HbwqrVYZE2SRUM2Jc5lMqlxaxX8Hdd109UKWofLsqxZVqLHxszXGngoBMzAbsNtdBrWBI+x29Xu8xtNJLDbGzWohbr3CrxrK8W3A4BtW9SYqqdEQg1fKCzT6+wSOY4bubjIFQD41g+ayLZL3hmrS1KSlZmbX4xciZKUiOYGMw/wfz/M0ldXnsgFGjzURV2sfRceF+3KhwPxCYdDQslhml+ImVq54KKlVwv6v7Pd9GFzXIA/f7Ui/T5qidE/Z3bT1MIyfXn5qfRBhb9/ptGmvv11dOLzFCVi0i3ULe560mVEBr/6lN4igW/+Nr5hRU+u8/TlfdlXqychr9QO8tQUTGcd+ul2TmT98EVK31nDtaamX1aWqukYxGpuaqB69nm4zfk/Xkizi0faqPfjFC29ydPCIoPUlH9H83T61gKlUzbdOt6qaaFnRa6AbhFDtOM4FRfxpzdmxNgj32X0aDXcyHchbJXYbTa1jTIa0502cgXUbLuBrqKZxCJrhffEgP2i2Kac2BYFyoWEfmc0pguyqhRwePusaQu4cW9bDW0z2hYLbDYWcmOmDkTRK6DbRsPGQuJC4SdfWm5bLDQ4sURdv07KBbpSLDT8f55c9oc0hxQmxciFCg7RtUdZ+v1ZnqbfOBIz+WMn0HePdhFCtpMjczSe4w6r9NixdprnffLA4CxzAlwlZlF/d530CXszUuDq4yzdfLqkgh+eKMDwLzuhA+ImAEIc5LCfl3YaNFajtNXZ1N+epFN7w8dBGNNIgd+m8gJgoKeFvs4k0H9daeZEDcrAZ61iBY/GcxfX3T8UPkxAn0rFKBb6ZHhW3OZBbzAI3SIGiDCxUCWwunojCzichPHZXzeefHFtOhbSjrgx00gBHDRhCgK6oLA5TH+TNkYKYAFWWzFSgOSory4VjBRQFgjnRjulptGxJ8FWiA9u7ET4tEy3NssFGrytlO9fLNLMynoXNlOAW1daB942iu/iGKdScIFWuaLNK/FnNk/fTr4kPP3FSIG2es7Gs9P99brgH2MN7eWFBl/lqOv+hFygK8VCawYJeIEFYAm/NYwOsh/ncnR9PldMo3hhgHpCqkWjB7uoPRkX4OMFlRfq5ETP2P4Omswv0557Y3IYKoywiAolpDe/+tNQHi1pm7KpznDcdDHdaBZKnNnVwPGMS78s5mlqhUMGBDUiX7mGCFKkwld+R/PVSwDzrQSf3ZPfMaQKRvCCrBEz+Mm/jaHLumJkgXW9NQwvLwS3OTmByJPjoo409bU0bJgX0gy1htX5RI0F5uFUUmYfVjDaRteMGPCCLDQSuQA81tJRbIVYCHVbKZ7bQAGvRK7AlvJCHirN0z/r/urIXcg/+E7QZWt7J0RGK+O9AtHOp/loHKHwfw9qtAC7zefDUI3i5wOOhmr/zx74ywr+9cE5nZ9rwZ2AEViBGdjfAhPs4mowdpbkAAAAAElFTkSuQmCC",dt=""+new URL("VirtualOutbound-3bb05b70.png",import.meta.url).href,yt={class:"policy-type-tag"},mt=["src"],ht=Y({__name:"PolicyTypeTag",props:{policyType:{type:String,required:!0}},setup(o){const l=o,B=te(),C={CircuitBreaker:{iconUrl:oe},FaultInjection:{iconUrl:ut},HealthCheck:{iconUrl:Ae},MeshAccessLog:{iconUrl:de},MeshCircuitBreaker:{iconUrl:oe},MeshGateway:{iconUrl:null},MeshGatewayRoute:{iconUrl:null},MeshHealthCheck:{iconUrl:Ae},MeshProxyPatch:{iconUrl:ce},MeshRateLimit:{iconUrl:re},MeshRetry:{iconUrl:ue},MeshTimeout:{iconUrl:pe},MeshTrace:{iconUrl:me},MeshTrafficPermission:{iconUrl:ye},ProxyTemplate:{iconUrl:ce},RateLimit:{iconUrl:re},Retry:{iconUrl:ue},Timeout:{iconUrl:pe},TrafficLog:{iconUrl:de},TrafficPermission:{iconUrl:ye},TrafficRoute:{iconUrl:pt},TrafficTrace:{iconUrl:me},VirtualOutbound:{iconUrl:dt}},D=x(()=>{const h=B.state.policyTypes.map(c=>{const g=C[c.name]??{iconUrl:null};return[c.name,g]});return Object.fromEntries(h)}),y=x(()=>D.value[l.policyType]);return(h,c)=>(e(),a("span",yt,[y.value.iconUrl!==null?(e(),a("img",{key:0,class:"policy-type-tag-icon",src:y.value.iconUrl,alt:""},null,8,mt)):(e(),d(z(Ue),{key:1,icon:"brain",size:"24"})),s(),ke(h.$slots,"default",{},()=>[s(v(l.policyType),1)],!0)]))}});const be=N(ht,[["__scopeId","data-v-0052ac03"]]),gt={class:"policy-type-heading"},ft={class:"policy-list"},vt={key:0,class:"origin-list"},Tt=Y({__name:"PolicyTypeEntryList",props:{id:{type:String,required:!1,default:"entry-list"},policyTypeEntries:{type:Object,required:!0}},setup(o){const l=o,B=[{label:"From",key:"sourceTags"},{label:"To",key:"destinationTags"},{label:"On",key:"name"},{label:"Conf",key:"config"},{label:"Origin policies",key:"origins"}];function C({headerKey:D}){return{class:`cell-${D}`}}return(D,y)=>{const h=J("router-link");return e(),d(V,{"initially-open":[],"multiple-open":""},{default:t(()=>[(e(!0),a(f,null,O(l.policyTypeEntries,(c,g)=>(e(),d(W,{key:g},{"accordion-header":t(()=>[T("h3",gt,[u(be,{"policy-type":c.type},{default:t(()=>[s(v(c.type)+" ("+v(c.connections.length)+`)
          `,1)]),_:2},1032,["policy-type"])])]),"accordion-content":t(()=>[T("div",ft,[u(z(fe),{class:"policy-type-table",fetcher:()=>({data:c.connections,total:c.connections.length}),headers:B,"cell-attrs":C,"disable-pagination":"","is-clickable":""},{sourceTags:t(({rowValue:n})=>[n.length>0?(e(),d(j,{key:0,class:"tag-list",tags:n},null,8,["tags"])):(e(),a(f,{key:1},[s(`
                —
              `)],64))]),destinationTags:t(({rowValue:n})=>[n.length>0?(e(),d(j,{key:0,class:"tag-list",tags:n},null,8,["tags"])):(e(),a(f,{key:1},[s(`
                —
              `)],64))]),name:t(({rowValue:n})=>[n!==null?(e(),a(f,{key:0},[s(v(n),1)],64)):(e(),a(f,{key:1},[s(`
                —
              `)],64))]),origins:t(({rowValue:n})=>[n.length>0?(e(),a("ul",vt,[(e(!0),a(f,null,O(n,(m,E)=>(e(),a("li",{key:`${g}-${E}`},[u(h,{to:m.route},{default:t(()=>[s(v(m.name),1)]),_:2},1032,["to"])]))),128))])):(e(),a(f,{key:1},[s(`
                —
              `)],64))]),config:t(({rowValue:n,rowKey:m})=>[n!==null?(e(),d(De,{key:0,id:`${l.id}-${g}-${m}-code-block`,code:n,language:"yaml","show-copy-button":!1},null,8,["id","code"])):(e(),a(f,{key:1},[s(`
                —
              `)],64))]),_:2},1032,["fetcher"])])]),_:2},1024))),128))]),_:1})}}});const Dt=N(Tt,[["__scopeId","data-v-2a8add12"]]),wt={class:"policy-type-heading"},Bt={class:"policy-list"},bt={key:1,class:"tag-list-wrapper"},Ct={key:0},Pt={key:1},kt={key:0,class:"list"},Et={key:0,class:"list"},Ot=Y({__name:"RuleEntryList",props:{id:{type:String,required:!1,default:"entry-list"},ruleEntries:{type:Object,required:!0}},setup(o){const l=o,B=[{label:"Type",key:"type"},{label:"Addresses",key:"addresses"},{label:"Conf",key:"config"},{label:"Origin policies",key:"origins"}];function C({headerKey:D}){return{class:`cell-${D}`}}return(D,y)=>{const h=J("router-link");return e(),d(V,{"initially-open":[],"multiple-open":""},{default:t(()=>[(e(!0),a(f,null,O(l.ruleEntries,(c,g)=>(e(),d(W,{key:g},{"accordion-header":t(()=>[T("h3",wt,[u(be,{"policy-type":c.type},{default:t(()=>[s(v(c.type)+" ("+v(c.connections.length)+`)
          `,1)]),_:2},1032,["policy-type"])])]),"accordion-content":t(()=>[T("div",Bt,[u(z(fe),{class:"policy-type-table",fetcher:()=>({data:c.connections,total:c.connections.length}),headers:B,"cell-attrs":C,"disable-pagination":"","is-clickable":""},{type:t(({rowValue:n})=>[n.sourceTags.length===0&&n.destinationTags.length===0?(e(),a(f,{key:0},[s(`
                —
              `)],64)):(e(),a("div",bt,[n.sourceTags.length>0?(e(),a("div",Ct,[s(`
                  From

                  `),u(j,{class:"tag-list",tags:n.sourceTags},null,8,["tags"])])):L("",!0),s(),n.destinationTags.length>0?(e(),a("div",Pt,[s(`
                  To

                  `),u(j,{class:"tag-list",tags:n.destinationTags},null,8,["tags"])])):L("",!0)]))]),addresses:t(({rowValue:n})=>[n.length>0?(e(),a("ul",kt,[(e(!0),a(f,null,O(n,(m,E)=>(e(),a("li",{key:`${g}-${E}`},v(m),1))),128))])):(e(),a(f,{key:1},[s(`
                —
              `)],64))]),origins:t(({rowValue:n})=>[n.length>0?(e(),a("ul",Et,[(e(!0),a(f,null,O(n,(m,E)=>(e(),a("li",{key:`${g}-${E}`},[u(h,{to:m.route},{default:t(()=>[s(v(m.name),1)]),_:2},1032,["to"])]))),128))])):(e(),a(f,{key:1},[s(`
                —
              `)],64))]),config:t(({rowValue:n,rowKey:m})=>[n!==null?(e(),d(De,{key:0,id:`${l.id}-${g}-${m}-code-block`,code:n,language:"yaml","show-copy-button":!1},null,8,["id","code"])):(e(),a(f,{key:1},[s(`
                —
              `)],64))]),_:2},1032,["fetcher"])])]),_:2},1024))),128))]),_:1})}}});const Qt=N(Ot,[["__scopeId","data-v-105d80e6"]]),Ce=o=>(he("data-v-ed201f38"),o=o(),ge(),o),Gt=Ce(()=>T("h2",{class:"visually-hidden"},`
    Policies
  `,-1)),Ut={key:0,class:"mt-2"},It=Ce(()=>T("h2",null,"Rules",-1)),Mt=Y({__name:"SidecarDataplanePolicyList",props:{dppName:{type:String,required:!0},policyTypeEntries:{type:Object,required:!0},ruleEntries:{type:Array,required:!0}},setup(o){const l=o;return(B,C)=>(e(),a(f,null,[Gt,s(),u(Dt,{id:"policies","policy-type-entries":l.policyTypeEntries},null,8,["policy-type-entries"]),s(),o.ruleEntries.length>0?(e(),a("div",Ut,[It,s(),u(Qt,{id:"rules","rule-entries":l.ruleEntries},null,8,["rule-entries"])])):L("",!0)],64))}});const Lt=N(Mt,[["__scopeId","data-v-ed201f38"]]),zt={key:2,class:"policies-list"},Rt={key:3,class:"policies-list"},xt=Y({__name:"DataplanePolicies",props:{dataPlane:{type:Object,required:!0}},setup(o){const l=o,B=ae(),C=te(),D=I(null),y=I([]),h=I([]),c=I([]),g=I([]),n=I(!0),m=I(null);Ee(()=>l.dataPlane.name,function(){E()}),E();async function E(){var p,A;m.value=null,n.value=!0,y.value=[],h.value=[],c.value=[],g.value=[];try{if(((A=(p=l.dataPlane.networking.gateway)==null?void 0:p.type)==null?void 0:A.toUpperCase())==="BUILTIN")D.value=await B.getMeshGatewayDataplane({mesh:l.dataPlane.mesh,name:l.dataPlane.name}),c.value=M(D.value),g.value=F(D.value.policies);else{const{items:r}=await B.getSidecarDataplanePolicies({mesh:l.dataPlane.mesh,name:l.dataPlane.name});y.value=X(r??[]);const{items:w}=await B.getDataplaneRules({mesh:l.dataPlane.mesh,name:l.dataPlane.name});h.value=Q(w??[])}}catch(i){i instanceof Error?m.value=i:console.error(i)}finally{n.value=!1}}function M(p){const A=[],i=p.listeners??[];for(const r of i)for(const w of r.hosts)for(const k of w.routes){const G=[];for(const U of k.destinations){const b=F(U.policies),S={routeName:k.route,route:{name:"policy-detail-view",params:{mesh:p.gateway.mesh,policyPath:"meshgatewayroutes",policy:k.route}},service:U.tags["kuma.io/service"],policies:b};G.push(S)}A.push({protocol:r.protocol,port:r.port,hostName:w.hostName,routeEntries:G})}return A}function F(p){if(p===void 0)return[];const A=[];for(const i of Object.values(p)){const r=C.state.policyTypesByName[i.type];A.push({type:i.type,name:i.name,route:{name:"policy-detail-view",params:{mesh:i.mesh,policyPath:r.path,policy:i.name}}})}return A}function X(p){const A=new Map;for(const r of p){const{type:w,service:k}=r,G=typeof k=="string"&&k!==""?[{label:"kuma.io/service",value:k}]:[],U=w==="inbound"||w==="outbound"?r.name:null;for(const[b,S]of Object.entries(r.matchedPolicies)){A.has(b)||A.set(b,{type:b,connections:[]});const _=A.get(b),H=C.state.policyTypesByName[b];for(const se of S){const R=Z(se,H,r,G,U);_.connections.push(...R)}}}const i=Array.from(A.values());return i.sort((r,w)=>r.type.localeCompare(w.type)),i}function Z(p,A,i,r,w){const k=p.conf&&Object.keys(p.conf).length>0?le(p.conf):null,U=[{name:p.name,route:{name:"policy-detail-view",params:{mesh:p.mesh,policyPath:A.path,policy:p.name}}}],b=[];if(i.type==="inbound"&&Array.isArray(p.sources))for(const{match:S}of p.sources){const H={sourceTags:[{label:"kuma.io/service",value:S["kuma.io/service"]}],destinationTags:r,name:w,config:k,origins:U};b.push(H)}else{const _={sourceTags:[],destinationTags:r,name:w,config:k,origins:U};b.push(_)}return b}function Q(p){const A=new Map;for(const r of p){A.has(r.policyType)||A.set(r.policyType,{type:r.policyType,connections:[]});const w=A.get(r.policyType),k=C.state.policyTypesByName[r.policyType],G=P(r,k);w.connections.push(...G)}const i=Array.from(A.values());return i.sort((r,w)=>r.type.localeCompare(w.type)),i}function P(p,A){const{type:i,service:r,subset:w,conf:k}=p,G=w?Object.entries(w):[];let U,b;i==="ClientSubset"?G.length>0?U=G.map(([R,$])=>({label:R,value:$})):U=[{label:"kuma.io/service",value:"*"}]:U=[],i==="DestinationSubset"?G.length>0?b=G.map(([R,$])=>({label:R,value:$})):typeof r=="string"&&r!==""?b=[{label:"kuma.io/service",value:r}]:b=[{label:"kuma.io/service",value:"*"}]:i==="ClientSubset"&&typeof r=="string"&&r!==""?b=[{label:"kuma.io/service",value:r}]:b=[];const S=p.addresses??[],_=k&&Object.keys(k).length>0?le(k):null,H=[];for(const R of p.origins)H.push({name:R.name,route:{name:"policy-detail-view",params:{mesh:R.mesh,policyPath:A.path,policy:R.name}}});return[{type:{sourceTags:U,destinationTags:b},addresses:S,config:_,origins:H}]}return(p,A)=>n.value?(e(),d(ve,{key:0})):m.value!==null?(e(),d(Be,{key:1,error:m.value},null,8,["error"])):y.value.length>0?(e(),a("div",zt,[u(Lt,{"dpp-name":o.dataPlane.name,"policy-type-entries":y.value,"rule-entries":h.value},null,8,["dpp-name","policy-type-entries","rule-entries"])])):c.value.length>0&&D.value!==null?(e(),a("div",Rt,[u(rt,{"mesh-gateway-dataplane":D.value,"mesh-gateway-listener-entries":c.value,"mesh-gateway-route-policies":g.value},null,8,["mesh-gateway-dataplane","mesh-gateway-listener-entries","mesh-gateway-route-policies"])])):(e(),d(we,{key:4}))}});const St=N(xt,[["__scopeId","data-v-bd1598bb"]]),Yt={class:"entity-heading"},Nt=["href"],_t=Y({__name:"DataPlaneDetails",props:{dataPlane:{type:Object,required:!0},dataPlaneOverview:{type:Object,required:!0}},setup(o){const l=o,{t:B}=Te(),C=ae(),D=te(),y=[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"DPP Insights"},{hash:"#dpp-policies",title:"Policies"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"},{hash:"#mtls",title:"Certificate Insights"},{hash:"#warnings",title:"Warnings"}],h=I([]),c=x(()=>({name:"data-plane-detail-view",params:{mesh:l.dataPlane.mesh,dataPlane:l.dataPlane.name}})),g=x(()=>ze(l.dataPlane,l.dataPlaneOverview.dataplaneInsight)),n=x(()=>ne(l.dataPlane)),m=x(()=>Re(l.dataPlaneOverview.dataplaneInsight)),E=x(()=>xe(l.dataPlaneOverview)),M=x(()=>{var P;const Q=Array.from(((P=l.dataPlaneOverview.dataplaneInsight)==null?void 0:P.subscriptions)??[]);return Q.reverse(),Q}),F=x(()=>h.value.length===0?y.filter(Q=>Q.hash!=="#warnings"):y);function X(){var A;const Q=((A=l.dataPlaneOverview.dataplaneInsight)==null?void 0:A.subscriptions)??[];if(Q.length===0||!("version"in Q[0]))return;const P=Q[0].version;if(P&&P.kumaDp&&P.envoy){const i=Se(P);i.kind!==Ye&&i.kind!==Ne&&h.value.push(i)}D.getters["config/getMulticlusterStatus"]&&P&&ne(l.dataPlane).find(w=>w.label===Oe)&&typeof P.kumaDp.kumaCpCompatible=="boolean"&&!P.kumaDp.kumaCpCompatible&&h.value.push({kind:_e,payload:{kumaDp:P.kumaDp.version}})}X();async function Z(Q){const{mesh:P,name:p}=l.dataPlane;return await C.getDataplaneFromMesh({mesh:P,name:p},Q)}return(Q,P)=>{const p=J("router-link");return e(),d(We,{tabs:F.value},{tabHeader:t(()=>[T("h1",Yt,[s(`
        DPP:

        `),u(Ke,{text:o.dataPlane.name},{default:t(()=>[u(p,{to:c.value},{default:t(()=>[s(v(o.dataPlane.name),1)]),_:1},8,["to"])]),_:1},8,["text"])])]),overview:t(()=>[u(ie,null,{default:t(()=>[n.value.length>0?(e(),d(K,{key:0,term:"Tags"},{default:t(()=>[u(j,{tags:n.value},null,8,["tags"])]),_:1})):L("",!0),s(),g.value.status?(e(),d(K,{key:1,term:"Status"},{default:t(()=>[u(Fe,{status:g.value.status},null,8,["status"])]),_:1})):L("",!0),s(),g.value.reason.length>0?(e(),d(K,{key:2,term:"Reason"},{default:t(()=>[(e(!0),a(f,null,O(g.value.reason,(A,i)=>(e(),a("div",{key:i,class:"reason"},v(A),1))),128))]),_:1})):L("",!0),s(),m.value!==null?(e(),d(K,{key:3,term:"Dependencies"},{default:t(()=>[T("ul",null,[(e(!0),a(f,null,O(m.value,(A,i)=>(e(),a("li",{key:i,class:"tag-cols"},v(i)+": "+v(A),1))),128))])]),_:1})):L("",!0)]),_:1}),s(),u(je,{id:"code-block-data-plane",class:"mt-4","resource-fetcher":Z,"resource-fetcher-watch-key":l.dataPlane.name,"is-searchable":""},null,8,["resource-fetcher-watch-key"])]),insights:t(()=>[u(Je,{"is-empty":M.value.length===0},{default:t(()=>[u(V,{"initially-open":0},{default:t(()=>[(e(!0),a(f,null,O(M.value,(A,i)=>(e(),d(W,{key:i},{"accordion-header":t(()=>[u(Me,{details:A},null,8,["details"])]),"accordion-content":t(()=>[u(Le,{details:A,"is-discovery-subscription":""},null,8,["details"])]),_:2},1024))),128))]),_:1})]),_:1},8,["is-empty"])]),"dpp-policies":t(()=>[u(St,{"data-plane":o.dataPlane},null,8,["data-plane"])]),"xds-configuration":t(()=>[u(ee,{"data-path":"xds",mesh:o.dataPlane.mesh,"dpp-name":o.dataPlane.name,"query-key":"envoy-data-data-plane"},null,8,["mesh","dpp-name"])]),"envoy-stats":t(()=>[u(ee,{"data-path":"stats",mesh:o.dataPlane.mesh,"dpp-name":o.dataPlane.name,"query-key":"envoy-data-data-plane"},null,8,["mesh","dpp-name"])]),"envoy-clusters":t(()=>[u(ee,{"data-path":"clusters",mesh:o.dataPlane.mesh,"dpp-name":o.dataPlane.name,"query-key":"envoy-data-data-plane"},null,8,["mesh","dpp-name"])]),mtls:t(()=>[E.value===null?(e(),d(z(Ie),{key:0,appearance:"danger"},{alertMessage:t(()=>[s(`
          This data plane proxy does not yet have mTLS configured —
          `),T("a",{href:z(B)("data-planes.href.docs.mutual-tls"),class:"external-link",target:"_blank"},`
            Learn About Certificates in `+v(z(B)("common.product.name")),9,Nt)]),_:1})):(e(),d(ie,{key:1},{default:t(()=>[(e(!0),a(f,null,O(E.value,(A,i)=>(e(),d(K,{key:i,term:z(B)(`http.api.property.${i}`)},{default:t(()=>[s(v(A),1)]),_:2},1032,["term"]))),128))]),_:1}))]),warnings:t(()=>[u(Ve,{warnings:h.value},null,8,["warnings"])]),_:1},8,["tabs"])}}});const Ht=N(_t,[["__scopeId","data-v-e7d30ff8"]]),qt={class:"kcard-border"},Aa=Y({__name:"DataPlaneDetailView",props:{isGatewayView:{type:Boolean,required:!1,default:!1}},setup(o){const l=o,B=ae(),C=Qe(),{t:D}=Te(),y=I(null),h=I(null),c=I(!0),g=I(null);n();async function n(){g.value=null,c.value=!0;const m=C.params.mesh,E=C.params.dataPlane;try{y.value=await B.getDataplaneFromMesh({mesh:m,name:E}),h.value=await B.getDataplaneOverviewFromMesh({mesh:m,name:E})}catch(M){y.value=null,M instanceof Error?g.value=M:console.error(M)}finally{c.value=!1}}return(m,E)=>(e(),d(qe,null,{default:t(({route:M})=>[u(Xe,{title:z(D)(`${l.isGatewayView?"gateways":"data-planes"}.routes.item.title`,{name:M.params.dataPlane})},null,8,["title"]),s(),u(He,{breadcrumbs:[{to:{name:`${l.isGatewayView?"gateways":"data-planes"}-list-view`,params:{mesh:M.params.mesh}},text:z(D)(`${l.isGatewayView?"gateways":"data-planes"}.routes.item.breadcrumbs`)}]},{default:t(()=>[T("div",qt,[c.value?(e(),d(ve,{key:0})):g.value!==null?(e(),d(Be,{key:1,error:g.value},null,8,["error"])):y.value===null||h.value===null?(e(),d(we,{key:2})):(e(),d(Ht,{key:3,"data-plane":y.value,"data-plane-overview":h.value},null,8,["data-plane","data-plane-overview"]))])]),_:2},1032,["breadcrumbs"])]),_:1}))}});export{Aa as default};
